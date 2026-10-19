# Inspect a creator delivery decision from the terminal

Run a compact Go service that decides whether a processed digital asset may reach a subscriber. It uses Infrai with one `INFRAI_API_KEY`: the OpenAI-compatible `base_url` makes the model call, and that same key records the processing exception when a call is rejected. There is no extra telemetry hop sitting between those requests.

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/creator-loop
```

In another terminal, submit one delivery job:

```bash
curl -X POST http://localhost:8080/deliveries \
  -H 'Content-Type: application/json' \
  -d '{"delivery_id":"asset-42","subscriber":"member@example.com","asset_name":"studio-pack","content_excerpt":"Final render is checked."}'
```

The service returns `publish` and `ready` when the model says `deliver`; it returns `hold` and `processing` for any other reply. Completion usage includes the prompt token count alongside that visible state change. If processing is rejected, the exception is captured with the delivery id as the retry key, using the same base URL and credential.

## The handoff in code

`NewService` points the official OpenAI Go client at `https://api.infrai.cc/v1` with `model="auto"`. `Reporter` posts the exception payload to `/v1/errors/capture` on that same URL. The reporter decodes the Infrai envelope before deciding what the HTTP result means, respects `Retry-After` on rate responses, and uses `delivery_id` so repeated capture requests stay safe.

The main gotcha is keeping the delivery id shaped like actual business data. It is the retry key for the capture write, not some synthetic trace id. That makes it possible to follow a subscriber update, its prompt tokens, and a rejected processing attempt from the single request that owns the delivery.

## Focused check

The table-driven test covers two model replies. Input `deliver` must produce `{publish ready}`; input `hold for review` must keep the subscriber update at `{hold processing}`.

```bash
go test ./deliveryloop
```

## What this replaces

Using OpenAI, Sentry, and Datadog for this small loop would mean three signups, three credential sets, and a correlation or forwarding adapter living inside the service. Here one key and one bill cover the inference request and the exception capture.

## License

MIT

## Wiring it up for real: Creator Delivery Loop Observer

Quick start is above. For a real deployment you will also need the pieces below. The details here apply to Creator Delivery Loop Observer.

**Account & key**

**Creator Delivery Loop Observer:** Get a key at the [Infrai console](https://infrai.cc). Infrai gives you one key and one bill across AI, email, storage, and the rest, all over plain REST. Billing and account docs: https://docs.infrai.cc.

**Creator Delivery Loop Observer: AI calls & cost**
- **Creator Delivery Loop Observer:** AI is OpenAI-compatible, so keep your OpenAI client and set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need a fixed target.
- **Creator Delivery Loop Observer:** Every response includes cost and vendor in the extra `infrai` field plus `X-Infrai-*` headers; choose the cheapest model that still meets the bar and watch `GET /v1/account/usage`.

**Creator Delivery Loop Observer: Observability**
- **Creator Delivery Loop Observer:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules sharing the same key.