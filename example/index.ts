import { serve } from "bun";
import { join, dirname } from "path";

const MOCKPAY_URL = process.env.MOCKPAY_URL || "http://localhost:8080";
const MOCKPAY_KEY = process.env.MOCKPAY_API_KEY || "mock_key";
const PORT = parseInt(process.env.PORT || "3000");

const ROOT = join(dirname(import.meta.dir), "example", "public");

function template(html: string) {
  return html
    .replace(/\{\{MOCKPAY_URL\}\}/g, MOCKPAY_URL)
    .replace(/\{\{MOCKPAY_KEY\}\}/g, MOCKPAY_KEY)
    .replace(/\{\{PORT\}\}/g, PORT.toString());
}

// SSE clients
const clients = new Set<ReadableStreamDefaultController>();

function broadcast(event: string, data: unknown) {
  const msg = `event: ${event}\ndata: ${JSON.stringify(data)}\n\n`;
  for (const c of clients) {
    try { c.enqueue(msg); } catch { clients.delete(c); }
  }
}

serve({
  port: PORT,
  async fetch(req) {
    const url = new URL(req.url);

    // Receive webhook from MockPay
    if (url.pathname === "/webhook" && req.method === "POST") {
      const body = await req.json();
      console.log("[webhook]", JSON.stringify(body));
      broadcast("webhook", body);
      return new Response("ok", { status: 200 });
    }

    // SSE stream for browser
    if (url.pathname === "/events") {
      const stream = new ReadableStream({
        start(controller) {
          clients.add(controller);
          controller.enqueue("event: connected\ndata: {}\n\n");
        },
        cancel(controller) {
          clients.delete(controller);
        },
      });
      return new Response(stream, {
        headers: {
          "Content-Type": "text/event-stream",
          "Cache-Control": "no-cache",
          Connection: "keep-alive",
        },
      });
    }

    // Static files
    const path = url.pathname === "/" ? "/index.html" : url.pathname;
    let filePath = join(ROOT, path);

    if (!path.includes(".")) {
      const htmlPath = filePath + ".html";
      if (await Bun.file(htmlPath).exists()) {
        filePath = htmlPath;
      }
    }

    const file = Bun.file(filePath);
    if (await file.exists()) {
      if (filePath.endsWith(".html")) {
        const html = await file.text();
        return new Response(template(html), {
          headers: { "Content-Type": "text/html; charset=utf-8" },
        });
      }
      return new Response(file);
    }

    return new Response("Not found", { status: 404 });
  },
});

console.log(`Playground running at http://localhost:${PORT}`);
console.log(`MockPay URL: ${MOCKPAY_URL}`);
console.log(`\n  Set MOCKPAY_WEBHOOK_URL=http://localhost:${PORT}/webhook when starting MockPay`);
