import request from "supertest";
import { createApp, resetState } from "../src/index";

const app = createApp();

beforeEach(() => {
  resetState();
});

describe("GET /health", () => {
  it("returns ok status", async () => {
    const res = await request(app).get("/health");
    expect(res.status).toBe(200);
    expect(res.body.status).toBe("ok");
    expect(res.body.service).toBe("dashboard-api");
  });
});

describe("POST /api/dashboards", () => {
  it("creates a new dashboard", async () => {
    const res = await request(app)
      .post("/api/dashboards")
      .send({ name: "My Dashboard", pipelineIds: ["p1", "p2"] });
    expect(res.status).toBe(201);
    expect(res.body.name).toBe("My Dashboard");
    expect(res.body.pipelineIds).toEqual(["p1", "p2"]);
    expect(res.body.id).toMatch(/^dash-/);
  });

  it("returns 400 when name is missing", async () => {
    const res = await request(app).post("/api/dashboards").send({});
    expect(res.status).toBe(400);
    expect(res.body.error).toContain("name");
  });

  it("defaults pipelineIds to empty array", async () => {
    const res = await request(app)
      .post("/api/dashboards")
      .send({ name: "Empty" });
    expect(res.status).toBe(201);
    expect(res.body.pipelineIds).toEqual([]);
  });
});

describe("GET /api/dashboards", () => {
  it("returns empty list initially", async () => {
    const res = await request(app).get("/api/dashboards");
    expect(res.status).toBe(200);
    expect(res.body).toEqual([]);
  });

  it("returns created dashboards", async () => {
    await request(app).post("/api/dashboards").send({ name: "D1" });
    await request(app).post("/api/dashboards").send({ name: "D2" });
    const res = await request(app).get("/api/dashboards");
    expect(res.status).toBe(200);
    expect(res.body).toHaveLength(2);
  });
});

describe("GET /api/dashboards/:id", () => {
  it("returns a specific dashboard", async () => {
    const created = await request(app)
      .post("/api/dashboards")
      .send({ name: "Test" });
    const res = await request(app).get(`/api/dashboards/${created.body.id}`);
    expect(res.status).toBe(200);
    expect(res.body.name).toBe("Test");
  });

  it("returns 404 for unknown id", async () => {
    const res = await request(app).get("/api/dashboards/unknown");
    expect(res.status).toBe(404);
  });
});

describe("DELETE /api/dashboards/:id", () => {
  it("deletes an existing dashboard", async () => {
    const created = await request(app)
      .post("/api/dashboards")
      .send({ name: "ToDelete" });
    const res = await request(app).delete(
      `/api/dashboards/${created.body.id}`
    );
    expect(res.status).toBe(204);
    const check = await request(app).get(
      `/api/dashboards/${created.body.id}`
    );
    expect(check.status).toBe(404);
  });

  it("returns 404 for unknown id", async () => {
    const res = await request(app).delete("/api/dashboards/nope");
    expect(res.status).toBe(404);
  });
});

describe("GET /api/overview", () => {
  it("returns overview stats", async () => {
    await request(app)
      .post("/api/dashboards")
      .send({ name: "D1", pipelineIds: ["p1", "p2"] });
    await request(app)
      .post("/api/dashboards")
      .send({ name: "D2", pipelineIds: ["p2", "p3"] });
    const res = await request(app).get("/api/overview");
    expect(res.status).toBe(200);
    expect(res.body.total_dashboards).toBe(2);
    expect(res.body.monitored_pipelines).toBe(3);
  });
});
