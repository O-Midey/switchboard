import { AppError } from "@/lib/errors";
import { getSwitchboardState, type SwitchboardState } from "@/lib/switchboard";

const fallback: SwitchboardState = { generatedAt: new Date(0).toISOString(), uptimeSeconds: 0, strategy: "offline", policySource: "unavailable", backends: [] };

export default async function Home() {
  let state = fallback;
  let unavailable = false;
  try { state = await getSwitchboardState(); } catch (error: unknown) { unavailable = error instanceof AppError; if (!unavailable) throw error; }
  const healthy = state.backends.filter((backend) => backend.healthy).length;
  const requests = state.backends.reduce((total, backend) => total + backend.metrics.requests, 0);
  return <main>
    <header className="masthead">
      <div className="brand"><span className="brand-mark" aria-hidden="true">S/B</span><div><strong>Switchboard</strong><span>Control room</span></div></div>
      <div className={`system-state ${unavailable ? "is-offline" : ""}`}><i />{unavailable ? "Control plane unreachable" : "Data plane observed"}</div>
    </header>
    <section className="hero">
      <div><p className="eyebrow">Routing posture</p><h1>Traffic, with<br/><em>judgment.</em></h1></div>
      <div className="routing-rail" aria-label={`${healthy} of ${state.backends.length} backends healthy`}>
        <div className="rail-source"><span>IN</span></div><div className="rail-line"><b /></div>
        <div className="rail-targets">{state.backends.length ? state.backends.map((backend) => <span className={backend.healthy ? "healthy" : "unhealthy"} key={backend.id}>{backend.id.slice(0, 1).toUpperCase()}</span>) : <span className="unhealthy">—</span>}</div>
      </div>
    </section>
    <section className="summary" aria-label="System summary">
      <article><span>Healthy targets</span><strong>{healthy}<small>/{state.backends.length}</small></strong></article>
      <article><span>Window requests</span><strong>{requests.toLocaleString()}</strong></article>
      <article><span>Routing rule</span><strong className="text-value">{state.strategy}</strong></article>
      <article><span>Policy authority</span><strong className="text-value">{state.policySource}</strong></article>
    </section>
    <section className="fleet">
      <div className="section-heading"><div><p className="eyebrow">Upstream fleet</p><h2>Live backend state</h2></div><p>Rolling 60-second window</p></div>
      <div className="backend-list">{state.backends.map((backend) => <article className="backend" key={backend.id}>
        <div className="backend-identity"><i className={backend.healthy ? "healthy" : "unhealthy"}/><div><h3>{backend.id}</h3><p>{backend.url}</p></div><span>{backend.healthy ? "Ready" : "Excluded"}</span></div>
        <dl><div><dt>Latency</dt><dd>{backend.metrics.avgLatencyMs.toFixed(1)}<small> ms</small></dd></div><div><dt>Error rate</dt><dd>{(backend.metrics.errorRate * 100).toFixed(1)}<small>%</small></dd></div><div><dt>Active</dt><dd>{backend.metrics.activeRequests}</dd></div><div><dt>Requests</dt><dd>{backend.metrics.requests}</dd></div></dl>
      </article>)}</div>
      {state.backends.length === 0 && <div className="empty"><strong>No telemetry yet.</strong><span>Start the local stack to connect this control room.</span></div>}
    </section>
    <footer><span>Deterministic data plane</span><span>Jev policy provider · disconnected by design</span></footer>
  </main>;
}
