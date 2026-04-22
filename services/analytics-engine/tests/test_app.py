import pytest
from app import create_app


@pytest.fixture
def client():
    app = create_app()
    app.config["TESTING"] = True
    with app.test_client() as c:
        yield c


@pytest.fixture(autouse=True)
def clear_pipelines():
    from app import pipelines
    pipelines.clear()


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["status"] == "ok"
    assert data["service"] == "analytics-engine"


def test_list_pipelines_empty(client):
    resp = client.get("/api/pipelines")
    assert resp.status_code == 200
    assert resp.get_json() == []


def test_create_pipeline(client):
    resp = client.post(
        "/api/pipelines",
        json={"id": "p1", "name": "Test Pipeline"},
    )
    assert resp.status_code == 201
    data = resp.get_json()
    assert data["id"] == "p1"
    assert data["name"] == "Test Pipeline"
    assert data["status"] == "active"
    assert data["events_count"] == 0


def test_create_pipeline_duplicate(client):
    client.post("/api/pipelines", json={"id": "p1", "name": "A"})
    resp = client.post("/api/pipelines", json={"id": "p1", "name": "B"})
    assert resp.status_code == 409


def test_create_pipeline_invalid(client):
    resp = client.post("/api/pipelines", json={"id": "p1"})
    assert resp.status_code == 400


def test_get_pipeline(client):
    client.post("/api/pipelines", json={"id": "p1", "name": "A"})
    resp = client.get("/api/pipelines/p1")
    assert resp.status_code == 200
    assert resp.get_json()["id"] == "p1"


def test_get_pipeline_not_found(client):
    resp = client.get("/api/pipelines/unknown")
    assert resp.status_code == 404


def test_ingest_event(client):
    client.post("/api/pipelines", json={"id": "p1", "name": "A"})
    resp = client.post("/api/pipelines/p1/ingest", json={"key": "value"})
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["accepted"] is True
    assert data["events_count"] == 1


def test_ingest_event_pipeline_not_found(client):
    resp = client.post("/api/pipelines/nope/ingest", json={"key": "v"})
    assert resp.status_code == 404


def test_ingest_event_invalid_body(client):
    client.post("/api/pipelines", json={"id": "p1", "name": "A"})
    resp = client.post(
        "/api/pipelines/p1/ingest",
        data="not json",
        content_type="text/plain",
    )
    assert resp.status_code == 400


def test_stats(client):
    client.post("/api/pipelines", json={"id": "p1", "name": "A"})
    client.post("/api/pipelines", json={"id": "p2", "name": "B"})
    client.post("/api/pipelines/p1/ingest", json={"x": 1})
    client.post("/api/pipelines/p1/ingest", json={"x": 2})
    resp = client.get("/api/stats")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["total_pipelines"] == 2
    assert data["active_pipelines"] == 2
    assert data["total_events"] == 2
