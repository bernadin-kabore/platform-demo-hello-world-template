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


def test_logs_are_json():
    """The collector's json_parser only promotes trace IDs from JSON bodies,
    so malformed log output silently breaks log/trace correlation."""
    import json as _json
    from app.main import JsonFormatter
    import logging as _logging

    record = _logging.LogRecord("t", _logging.INFO, __file__, 1, "hello", None, None)
    payload = _json.loads(JsonFormatter().format(record))
    assert payload["message"] == "hello"
    assert payload["severity"] == "INFO"
    assert "timestamp" in payload
