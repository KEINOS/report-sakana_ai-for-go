# Sakana AI API Notes

This file is a handoff note for AI agents that work on this repository.

## Scope

- Provider: Sakana AI
- Base URL: `https://api.sakana.ai/v1`
- Model used by this research: `fugu`
- API shape used by this research: OpenAI-compatible Chat Completions
- Subscription used for the benchmark: Standard

## Authentication

The access token is stored in the `SAKANA_AI_API_KEY` environment variable.
Do not confuse it with the variable name in the official examples, `SAKANA_API_KEY`.

Never print, log, commit, or send the token to a review worker.

## Available models

The Models API is:

```text
GET https://api.sakana.ai/v1/models
```

The IDs confirmed on June 23, 2026 were:

- `fugu`
- `fugu-ultra`
- `fugu-ultra-20260615`

This research must use only `fugu`.

## Chat Completions

Endpoint:

```text
POST https://api.sakana.ai/v1/chat/completions
```

Important request fields:

| Field | Note |
| :---: | :--- |
| `model` | Required. Use `fugu`. |
| `messages` | Required. OpenAI-compatible message array. |
| `stream` | Boolean. This research uses `false`. |
| `max_completion_tokens` | Preferred output limit field. |
| `max_tokens` | Legacy output limit field. |
| `reasoning_effort` | Accepts `high`, `xhigh`, or `max`. |
| `temperature` | Accepted but ignored by Sakana AI. |
| `top_p`, `stop`, `seed` | Accepted but ignored. |

Chat Completions defaults to `high` reasoning effort. The benchmark leaves that default unchanged.

The benchmark sent `max_completion_tokens: 10000`, but the API usage records reported up to 12,683 completion tokens with `finish_reason: stop`. For Fugu, do not treat this request value as a strict cap on the total usage reported for the complete orchestrated request.

## Connection check

```sh
curl -X POST https://api.sakana.ai/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${SAKANA_AI_API_KEY}" \
  -d '{"model":"fugu","messages":[{"role":"user","content":"How many r in word strawberry"}]}'
```

## Benchmark request

The benchmark sends one system message and one user task:

```json
{
  "model": "fugu",
  "messages": [
    {
      "role": "system",
      "content": "You are an expert Go engineer. Follow the requested API exactly. Output source code only."
    },
    {
      "role": "user",
      "content": "<task prompt>"
    }
  ],
  "temperature": 0,
  "max_completion_tokens": 10000,
  "stream": false
}
```

`temperature: 0` is kept to match the Sakura AI Engine benchmark request. But it does not make Fugu deterministic because Sakana AI ignores this field and `fugu` dynamically routes work to underlying models.

## Pricing note

Fugu has no single fixed token price. With pay-as-you-go billing, one active agent uses the selected underlying model's rate. For multiple agents, Sakana AI charges one rate based on the highest-tier model used.

The Standard subscription is USD 20 per month with a baseline allowance as of June 23, 2026.

For this reason, the final report must not invent a fixed per-token Fugu cost or directly include Fugu in the Sakura fixed-price Pareto calculation.

## Primary sources

- <https://console.sakana.ai/get-started>
- <https://console.sakana.ai/models>
- <https://console.sakana.ai/pricing>
