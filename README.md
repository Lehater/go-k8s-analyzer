# go-k8s-analyzer

Сервис на Go для приёма потоковых метрик, их статистического анализа (rolling average + z-score) и экспорта Prometheus-метрик, рассчитанный на развёртывание в Kubernetes.

## Функциональность

- HTTP API:
  - `POST /ingest` — приём метрики в формате JSON:
    - `timestamp` (RFC3339, опционально),
    - `cpu` (0–100),
    - `rps` (RPS, неотрицательное).
  - `GET /analyze` — текущая статистика по RPS (rolling окно и z-score, аномалия при `|z| > 2`).
  - `GET /metrics` — Prometheus-метрики (`http_requests_total`, `http_request_duration_seconds`, `anomalies_total`).
  - `GET /healthz` — health-check.
  - `GET /debug/pprof/*` — базовое профилирование pprof.
- Хранение:
  - Кэширование метрик в Redis с TTL (по key = timestamp).
- Аналитика:
  - Окно скользящей статистики (по умолчанию 50 событий, настраивается `ANALYTICS_WINDOW_SIZE`).
  - Подсчёт среднего, стандартного отклонения, z-score и счётчика аномалий.
- Метрики:
  - HTTP RPS, латентность, число аномалий — экспорт для Prometheus/Grafana.

## Локальный запуск без Kubernetes

Требования:

- Go 1.22+
- Redis (локальный или в Docker)

Пример:

```bash
docker run -d --name redis -p 6379:6379 redis:7-alpine

export REDIS_ADDR=localhost:6379
go run ./cmd/server
```

Проверка:

```bash
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -d '{"cpu": 10, "rps": 100}'

curl http://localhost:8080/analyze
curl http://localhost:8080/metrics
```

## Сборка Docker-образа

```bash
docker build -t go-k8s-analyzer:latest .
```

Образ multi-stage, на базе `golang:1.22-alpine` → `alpine:3.20`, размер < 300MB.

## Развёртывание в Minikube

1. Запуск Minikube:

```bash
minikube start --cpus=2 --memory=4096
```

2. Использование локального Docker внутри Minikube:

```bash
eval "$(minikube docker-env)"
docker build -t go-k8s-analyzer:latest .
```

3. Redis (можно оставить простой Deployment/Service или заменить на Helm-чарт `bitnami/redis`):

```bash
kubectl apply -f deployments/k8s-redis.yaml
```

4. Go-сервис, Service, Ingress, HPA:

```bash
kubectl apply -f deployments/k8s-go-deployment.yaml
kubectl apply -f deployments/k8s-go-service.yaml
kubectl apply -f deployments/k8s-ingress.yaml
kubectl apply -f deployments/k8s-go-hpa.yaml
```

5. Ingress (NGINX):

- В Minikube:

```bash
minikube addons enable ingress
```

- После этого `go-analyzer.local` можно добавить в `/etc/hosts`, привязав к IP Minikube:

```bash
echo "$(minikube ip) go-analyzer.local" | sudo tee -a /etc/hosts
```

Эндпоинты:

- `http://go-analyzer.local/analyze`
- `http://go-analyzer.local/metrics`

## Prometheus и Grafana в Kubernetes (Helm)

Пример через Helm (в кластере должны быть установлены `helm` и доступ в интернет):

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

helm install prometheus prometheus-community/prometheus \
  -f deployments/prometheus-values.yaml

helm install grafana grafana/grafana \
  -f deployments/grafana-values.yaml
```

Alertmanager можно настроить с помощью `deployments/alert-rules.yaml` (правило на >5 аномалий в минуту).

В Grafana создайте дашборды:

- RPS: `rate(http_requests_total[1m])`
- Latency: `histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))`
- Anomaly rate: `increase(anomalies_total[1m])`

## Нагрузочное тестирование (Locust)

Для генерации нагрузки (500–1000+ RPS) можно использовать Locust.

1. Установите Locust:

```bash
pip install locust
```

2. Используйте `tests/locustfile.py`. Пример запуска против локального сервера:

```bash
locust -f tests/locustfile.py --headless \
  -u 1000 -r 100 \
  --run-time 5m \
  --host http://localhost:8080
```

Параметры:

- `-u` — количество одновременных пользователей (подбирается для достижения 1000 RPS).
- `-r` — скорость разгона.
- `--run-time` — время теста (например, 5 минут).

Для кластера в Kubernetes можно использовать `--host http://go-analyzer.local`.

## Проверка аналитики (rolling average и z-score)

- Синтетические данные можно подавать через Locust или небольшой скрипт, меняя RPS:
  - нормальный режим: RPS в диапазоне 100–300;
  - аномалии: резкие скачки до 800–1000 RPS.
- Через `/analyze` вы получаете:
  - `mean`, `std_dev`, `z_score`, `is_anomaly`, `anomaly_count`.
- Для отчёта:
  - измерьте долю корректно найденных аномалий (TP / (TP+FN));
  - измерьте долю ложных срабатываний (FP / (FP+TN));
  - стремитесь к точности >70% и false positive <10% на вашем синтетическом сценарии.

## Профилирование и оптимизация

- Профилирование CPU/heap:

```bash
go tool pprof http://localhost:8080/debug/pprof/profile
```

- Оптимизации:
  - подстройка размера буфера `INGEST_BUFFER_SIZE`;
  - настройка Redis (например, `maxmemory-policy` и лимиты памяти);
  - анализ латентности и GC по pprof.

## Killercoda / облако

- В Killercoda или облаке (Yandex.Cloud / VK Cloud) шаги аналогичны:
  - запуск кластера Kubernetes;
  - установка Helm (если нужно);
  - развёртывание Redis/Go-сервиса/Prometheus/Grafana теми же манифестами и chart-ами.
- Локальные команды Minikube имеют аналоги:
  - `kubectl apply` — те же;
  - `minikube ip` — можно заменить на внешний адрес ingress-controller’а или LoadBalancer.

Для отчёта и презентации используйте:

- скриншоты `kubectl get pods`, дашбордов Grafana, логов Locust;
- YAML-файлы из директории `deployments/`;
- исходный код из `cmd/` и `internal/`.
