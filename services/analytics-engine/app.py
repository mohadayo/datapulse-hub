import logging
import os
import time
from flask import Flask, jsonify, request

app = Flask(__name__)

LOG_LEVEL = os.environ.get("LOG_LEVEL", "INFO").upper()
logging.basicConfig(
    level=getattr(logging, LOG_LEVEL, logging.INFO),
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
)
logger = logging.getLogger("analytics-engine")

pipelines: dict[str, dict] = {}


@app.route("/health")
def health():
    return jsonify({"status": "ok", "service": "analytics-engine"})


@app.route("/api/pipelines", methods=["GET"])
def list_pipelines():
    logger.info("Listing all pipelines")
    return jsonify(list(pipelines.values()))


@app.route("/api/pipelines/<pipeline_id>", methods=["GET"])
def get_pipeline(pipeline_id: str):
    pipeline = pipelines.get(pipeline_id)
    if not pipeline:
        logger.warning("Pipeline not found: %s", pipeline_id)
        return jsonify({"error": f"Pipeline '{pipeline_id}' not found"}), 404
    return jsonify(pipeline)


@app.route("/api/pipelines", methods=["POST"])
def create_pipeline():
    data = request.get_json(silent=True)
    if not data or "id" not in data or "name" not in data:
        logger.error("Invalid pipeline data: %s", data)
        return jsonify({"error": "Fields 'id' and 'name' are required"}), 400
    pipeline_id = data["id"]
    if pipeline_id in pipelines:
        return jsonify({"error": f"Pipeline '{pipeline_id}' already exists"}), 409
    pipeline = {
        "id": pipeline_id,
        "name": data["name"],
        "status": "active",
        "events_count": 0,
        "created_at": time.time(),
    }
    pipelines[pipeline_id] = pipeline
    logger.info("Created pipeline: %s", pipeline_id)
    return jsonify(pipeline), 201


@app.route("/api/pipelines/<pipeline_id>/ingest", methods=["POST"])
def ingest_event(pipeline_id: str):
    pipeline = pipelines.get(pipeline_id)
    if not pipeline:
        return jsonify({"error": f"Pipeline '{pipeline_id}' not found"}), 404
    data = request.get_json(silent=True)
    if not data:
        return jsonify({"error": "Request body must be valid JSON"}), 400
    pipeline["events_count"] += 1
    pipeline["last_event_at"] = time.time()
    logger.info(
        "Ingested event for pipeline %s (total: %d)",
        pipeline_id,
        pipeline["events_count"],
    )
    return jsonify({"accepted": True, "events_count": pipeline["events_count"]})


@app.route("/api/stats", methods=["GET"])
def stats():
    total_events = sum(p["events_count"] for p in pipelines.values())
    active = sum(1 for p in pipelines.values() if p["status"] == "active")
    return jsonify({
        "total_pipelines": len(pipelines),
        "active_pipelines": active,
        "total_events": total_events,
    })


def create_app():
    return app


if __name__ == "__main__":
    port = int(os.environ.get("ANALYTICS_PORT", "8001"))
    logger.info("Starting analytics-engine on port %d", port)
    app.run(host="0.0.0.0", port=port)
