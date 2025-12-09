from locust import HttpUser, task, between
import json
import time
import random


class MetricUser(HttpUser):
    wait_time = between(0.0, 0.01)

    @task
    def send_metric(self):
        # Имитация сценария как в синтетическом тесте:
        # большинство точек вокруг ~200 RPS, каждая 10-я — аномальный всплеск ~350 RPS.
        if not hasattr(self, "counter"):
            self.counter = 0
        self.counter += 1

        now = time.time()
        cpu = random.uniform(5, 80)

        if self.counter % 10 == 0:
            # аномальный всплеск нагрузки
            rps = random.uniform(330, 360)
        else:
            # нормальный режим вокруг 200 RPS
            rps = random.uniform(195, 205)

        payload = {
            "cpu": cpu,
            "rps": rps,
        }
        self.client.post(
            "/ingest",
            data=json.dumps(payload),
            headers={"Content-Type": "application/json"},
            name="/ingest",
        )
