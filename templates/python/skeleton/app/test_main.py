from fastapi.testclient import TestClient

from app.main import app

client = TestClient(app)


def test_hello():
    response = client.get("/")
    assert response.status_code == 200
    assert "message" in response.json()


def test_healthz():
    assert client.get("/healthz").status_code == 200


def test_readyz():
    assert client.get("/readyz").status_code == 200


def test_metrics():
    response = client.get("/metrics")
    assert response.status_code == 200
