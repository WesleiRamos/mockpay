const getUrl = () => document.getElementById("cfg-url").value.replace(/\/$/, "");
const getKey = () => document.getElementById("cfg-key").value;

async function api(method, path, body) {
  const r = await fetch(getUrl() + path, {
    method,
    headers: {
      Authorization: "Bearer " + getKey(),
      "Content-Type": "application/json",
    },
    body: body ? JSON.stringify(body) : undefined,
  });
  return r.json();
}

function formatCents(c) {
  return "R$ " + (c / 100).toFixed(2).replace(".", ",");
}

function showResponse(data) {
  const el = document.getElementById("response");
  const lbl = document.getElementById("response-label");
  if (!el || !lbl) return;

  const json = JSON.stringify(data, null, 2);
  const safe = json
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
  const highlighted = safe
    .replace(/"([^"]+)":/g, '<span style="color:#9fe870">"$1"</span>:')
    .replace(/: "([^"]*)"/g, ': <span style="color:#a8d8f0">"$1"</span>')
    .replace(/: (\d+\.?\d*)/g, ': <span style="color:#f5c842">$1</span>')
    .replace(/: (true|false)/g, ': <span style="color:#f0a500">$1</span>')
    .replace(/: (null)/g, ': <span style="color:#868685">$1</span>');

  el.innerHTML = highlighted;
  el.className = "visible";
  lbl.style.display = "block";
  el.scrollIntoView({ behavior: "smooth", block: "nearest" });
}

function clearResponse() {
  const el = document.getElementById("response");
  const lbl = document.getElementById("response-label");
  if (el)  { el.className = ""; el.innerHTML = ""; }
  if (lbl) lbl.style.display = "none";
}

function showResults(containerId, items, type) {
  let el = document.getElementById(containerId + "-list");
  if (!el) {
    el = document.createElement("div");
    el.id = containerId + "-list";
    el.style.cssText = "margin-top:16px;display:flex;flex-direction:column;gap:10px;";
    const panel = document.getElementById("panel-" + containerId);
    if (panel) panel.appendChild(el);
  }
  el.innerHTML = "";

  if (!items || !items.length) {
    el.innerHTML = '<p style="font-size:13px;color:var(--gray);font-weight:500;">No items found.</p>';
    return;
  }

  items.forEach((item) => {
    const card = document.createElement("div");
    card.dataset.id = item.id || "";
    card.style.cssText = "display:flex;justify-content:space-between;align-items:center;background:var(--white);border-radius:14px;padding:14px 18px;";

    const left = document.createElement("div");
    left.style.cssText = "display:flex;flex-direction:column;gap:3px;";

    const idEl = document.createElement("span");
    idEl.style.cssText = "font-size:11px;font-weight:500;color:var(--gray);font-variant-numeric:tabular-nums;";
    idEl.textContent = item.id || "";

    const amtEl = document.createElement("span");
    amtEl.style.cssText = "font-size:16px;font-weight:900;letter-spacing:-0.03em;color:var(--near-black);";
    amtEl.textContent = item.amount ? formatCents(item.amount) : "";

    const metaEl = document.createElement("span");
    metaEl.style.cssText = "font-size:12px;font-weight:500;color:var(--warm-dark);";
    const meta = type === "billing" ? (item.methods || []).join(", ") : "PIX";
    metaEl.textContent = meta + (item.frequency ? " · " + item.frequency : "");

    left.appendChild(idEl);
    if (item.amount) left.appendChild(amtEl);
    if (type !== "customer") left.appendChild(metaEl);

    const right = document.createElement("div");
    right.style.cssText = "display:flex;flex-direction:column;align-items:flex-end;gap:8px;";

    if (item.status && type !== "customer") {
      const badge = document.createElement("span");
      badge.dataset.statusTarget = item.id;
      badge.style.cssText = statusStyle(item.status);
      badge.textContent = item.status;
      right.appendChild(badge);
    }

    if (item.id && type !== "customer") {
      const link = document.createElement("a");
      link.href = getUrl() + "/checkout/" + item.id;
      link.target = "_blank";
      link.style.cssText = "font-size:12px;font-weight:700;color:var(--near-black);text-decoration:none;padding:5px 12px;border-radius:9999px;background:var(--bg);transition:transform .18s cubic-bezier(.34,1.56,.64,1);";
      link.textContent = "Checkout →";
      link.onmouseenter = () => link.style.transform = "scale(1.06)";
      link.onmouseleave = () => link.style.transform = "";
      right.appendChild(link);
    }

    card.appendChild(left);
    card.appendChild(right);
    el.appendChild(card);
  });
}

function statusStyle(status) {
  const base = "display:inline-flex;align-items:center;gap:4px;padding:3px 10px;border-radius:9999px;font-size:10px;font-weight:700;letter-spacing:.07em;text-transform:uppercase;";
  const map = {
    PENDING:  base + "background:rgba(255,209,26,.18);color:#7a5500;",
    APPROVED: base + "background:var(--light-mint);color:var(--pos-green);",
    DENIED:   base + "background:rgba(208,50,56,.09);color:var(--danger-red);",
    EXPIRED:  base + "background:rgba(14,15,12,.07);color:var(--warm-dark);",
    ACTIVE:   base + "background:var(--light-mint);color:var(--pos-green);",
  };
  return map[status] || map.PENDING;
}

function switchPanel(name) {
  document.querySelectorAll(".panel").forEach((p) => p.classList.remove("active"));
  document.querySelectorAll(".sidebar a").forEach((a) => a.classList.remove("active"));
  document.getElementById("panel-" + name).classList.add("active");
  document.getElementById("nav-" + name).classList.add("active");
  clearResponse();
}

function initWebhooks() {
  const es = new EventSource("/events");
  es.addEventListener("webhook", (e) => {
    const event = JSON.parse(e.data);
    const payload = event.payload;
    if (!payload || !payload.id) return;
    document.querySelectorAll("[data-status-target='" + payload.id + "']").forEach((badge) => {
      badge.style.cssText = statusStyle(payload.status);
      badge.textContent = payload.status;
    });
  });
}

async function checkHealth() {
  const r = await fetch(getUrl() + "/health").then((r) => r.json());
  showResponse(r);
}

async function createCustomer() {
  const r = await api("POST", "/v1/customer/create", {
    name:      document.getElementById("c-name").value,
    email:     document.getElementById("c-email").value,
    cellphone: document.getElementById("c-phone").value,
    tax_id:    document.getElementById("c-tax").value,
  });
  showResponse(r);
}

async function listCustomers() {
  const r = await api("GET", "/v1/customer/list");
  showResponse(r);
}

async function createCoupon() {
  const r = await api("POST", "/v1/coupon/create", {
    code:          document.getElementById("cp-code").value,
    discount_kind: document.getElementById("cp-kind").value,
    discount:      parseInt(document.getElementById("cp-discount").value),
    max_redeems:   parseInt(document.getElementById("cp-max").value),
  });
  showResponse(r);
}

async function listCoupons() {
  const r = await api("GET", "/v1/coupon/list");
  showResponse(r);
}

let prodCounter = 0;

function addProduct(name, qty, price) {
  const list = document.getElementById("products-list");
  const row = document.createElement("div");
  row.className = "product-row";

  const n = document.createElement("input");
  n.placeholder = "Product name";
  n.value = name || "";

  const q = document.createElement("input");
  q.type = "number"; q.placeholder = "Qty";
  q.value = String(qty || 1); q.min = "1";

  const p = document.createElement("input");
  p.type = "number"; p.placeholder = "Price (¢)";
  p.value = String(price || 5000); p.min = "100";

  const rm = document.createElement("button");
  rm.className = "btn-remove";
  rm.textContent = "×";
  rm.onclick = () => row.remove();

  // order: name → qty → price → remove
  row.appendChild(n);
  row.appendChild(q);
  row.appendChild(p);
  row.appendChild(rm);
  list.appendChild(row);
}

function getSelectedMethods() {
  const methods = [];
  document.querySelectorAll(".method-checkbox:checked").forEach((cb) => {
    methods.push(cb.value);
  });
  return methods;
}

async function createBilling() {
  const methods = getSelectedMethods();

  const prods = [];
  document.querySelectorAll("#products-list .product-row").forEach((r) => {
    const inputs = r.querySelectorAll("input");
    prods.push({
      external_id: "p" + ++prodCounter,
      name:        inputs[0].value,
      quantity:    parseInt(inputs[1].value) || 1,
      price:       parseInt(inputs[2].value) || 100,
    });
  });

  const body = {
    frequency:      document.getElementById("b-freq").value,
    methods,
    products:       prods,
    installments:   parseInt(document.getElementById("b-inst").value) || 1,
    interest_rate:  parseFloat(document.getElementById("b-rate").value) || 0,
    return_url:     document.getElementById("b-return").value,
    completion_url: document.getElementById("b-done").value,
  };

  const coupon = document.getElementById("b-coupon").value;
  if (coupon) body.coupon_code = coupon;

  const eid = document.getElementById("b-external-id").value;
  if (eid) body.external_id = eid;

  const ce = document.getElementById("b-cust-email").value;
  const cn = document.getElementById("b-cust-name").value;
  if (ce) body.customer = { email: ce, name: cn };

  const r = await api("POST", "/v1/billing/create", body);
  showResponse(r);
  if (r.data) showResults("billing", [r.data], "billing");
}

async function createPix() {
  const body = {
    amount:     parseInt(document.getElementById("px-amount").value),
    expires_in: parseInt(document.getElementById("px-expires").value),
  };

  const desc = document.getElementById("px-desc").value;
  if (desc) body.description = desc;

  const eid = document.getElementById("px-external-id").value;
  if (eid) body.external_id = eid;

  const ce = document.getElementById("px-cust-email").value;
  const cn = document.getElementById("px-cust-name").value;
  if (ce) body.customer = { email: ce, name: cn };

  const r = await api("POST", "/v1/pix/create", body);
  showResponse(r);

  if (r.data) {
    showResults("pix", [r.data], "pix");

    if (r.data.br_code_base64) {
      let qr = document.getElementById("pix-qr");
      if (!qr) {
        qr = document.createElement("div");
        qr.id = "pix-qr";
        qr.style.cssText = "margin-top:20px;display:flex;flex-direction:column;align-items:flex-start;gap:12px;";
        document.getElementById("panel-pix").appendChild(qr);
      }
      qr.innerHTML = "";

      const img = document.createElement("img");
      img.src = r.data.br_code_base64;
      img.style.cssText = "width:160px;height:160px;border-radius:14px;background:var(--white);padding:8px;";

      const a = document.createElement("a");
      a.href = getUrl() + "/checkout/" + r.data.id;
      a.target = "_blank";
      a.style.cssText = "font-size:13px;font-weight:700;color:var(--dark-green);text-decoration:none;padding:8px 18px;border-radius:9999px;background:var(--mockpay-green);transition:transform .18s cubic-bezier(.34,1.56,.64,1);display:inline-block;";
      a.textContent = "Open checkout →";
      a.onmouseenter = () => a.style.transform = "scale(1.05)";
      a.onmouseleave = () => a.style.transform = "";

      qr.appendChild(img);
      qr.appendChild(a);
    }
  }
}

async function loadStats() {
  const r = await api("GET", "/v1/stats");
  showResponse(r);

  const grid = document.getElementById("stats-grid");
  if (!grid || !r.data) return;

  const d = r.data;
  const blocks = [
    { label: "Customers",         value: d.customers_total      ?? 0, cls: "" },
    { label: "Coupons",           value: d.coupons_total         ?? 0, cls: "" },
    { label: "Billings total",    value: d.billings_total        ?? 0, cls: "" },
    { label: "Billings approved", value: d.billings_approved     ?? 0, cls: "val-green" },
    { label: "Billings pending",  value: d.billings_pending      ?? 0, cls: "val-yellow" },
    { label: "Billings denied",   value: d.billings_denied       ?? 0, cls: "val-red" },
    { label: "PIX total",         value: d.pix_total             ?? 0, cls: "" },
    { label: "PIX approved",      value: d.pix_approved          ?? 0, cls: "val-green" },
    { label: "PIX pending",       value: d.pix_pending           ?? 0, cls: "val-yellow" },
    { label: "PIX expired",       value: d.pix_expired           ?? 0, cls: "val-red" },
  ];

  grid.innerHTML = "";
  grid.style.display = "grid";

  blocks.forEach(({ label, value, cls }) => {
    const block = document.createElement("div");
    block.className = "stat-block";

    const lbl = document.createElement("div");
    lbl.className = "lbl";
    lbl.textContent = label;

    const val = document.createElement("div");
    val.className = "val " + cls;
    val.textContent = value;

    block.appendChild(lbl);
    block.appendChild(val);
    grid.appendChild(block);
  });
}

addProduct("T-shirt", 1, 5000);
addProduct("Cap", 1, 3000);
initWebhooks();
