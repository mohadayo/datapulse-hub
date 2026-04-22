import express, { Request, Response } from "express";

const app = express();
app.use(express.json());

interface Dashboard {
  id: string;
  name: string;
  pipelineIds: string[];
  createdAt: string;
}

const dashboards = new Map<string, Dashboard>();
let nextId = 0;

function log(level: string, message: string): void {
  const ts = new Date().toISOString();
  const logLevel = (process.env.LOG_LEVEL || "INFO").toUpperCase();
  const levels = ["DEBUG", "INFO", "WARN", "ERROR"];
  if (levels.indexOf(level) >= levels.indexOf(logLevel)) {
    console.log(`${ts} [${level}] dashboard-api: ${message}`);
  }
}

app.get("/health", (_req: Request, res: Response) => {
  res.json({ status: "ok", service: "dashboard-api" });
});

app.get("/api/dashboards", (_req: Request, res: Response) => {
  log("INFO", `Listing ${dashboards.size} dashboards`);
  res.json(Array.from(dashboards.values()));
});

app.get("/api/dashboards/:id", (req: Request<{ id: string }>, res: Response) => {
  const dashboard = dashboards.get(req.params.id);
  if (!dashboard) {
    log("WARN", `Dashboard not found: ${req.params.id}`);
    res.status(404).json({ error: `Dashboard '${req.params.id}' not found` });
    return;
  }
  res.json(dashboard);
});

app.post("/api/dashboards", (req: Request, res: Response) => {
  const { name, pipelineIds } = req.body;
  if (!name || typeof name !== "string") {
    log("ERROR", "Invalid dashboard data: name is required");
    res.status(400).json({ error: "Field 'name' is required and must be a string" });
    return;
  }
  nextId++;
  const id = `dash-${nextId}`;
  const dashboard: Dashboard = {
    id,
    name,
    pipelineIds: Array.isArray(pipelineIds) ? pipelineIds : [],
    createdAt: new Date().toISOString(),
  };
  dashboards.set(id, dashboard);
  log("INFO", `Created dashboard: ${id}`);
  res.status(201).json(dashboard);
});

app.delete("/api/dashboards/:id", (req: Request<{ id: string }>, res: Response) => {
  if (!dashboards.has(req.params.id)) {
    res.status(404).json({ error: `Dashboard '${req.params.id}' not found` });
    return;
  }
  dashboards.delete(req.params.id);
  log("INFO", `Deleted dashboard: ${req.params.id}`);
  res.status(204).send();
});

app.get("/api/overview", (_req: Request, res: Response) => {
  const allPipelineIds = new Set<string>();
  dashboards.forEach((d) => d.pipelineIds.forEach((p) => allPipelineIds.add(p)));
  res.json({
    total_dashboards: dashboards.size,
    monitored_pipelines: allPipelineIds.size,
  });
});

export function createApp(): express.Application {
  return app;
}

export function resetState(): void {
  dashboards.clear();
  nextId = 0;
}

if (require.main === module) {
  const port = parseInt(process.env.DASHBOARD_PORT || "8003", 10);
  app.listen(port, "0.0.0.0", () => {
    log("INFO", `Starting dashboard-api on port ${port}`);
  });
}
