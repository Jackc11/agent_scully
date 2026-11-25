### 1. Shared Sampling & Generation Parameters

These control randomness, token selection, and length in both engines:

| Parameter (Ollama / vLLM) | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `temperature` | float | `0.8` (Ollama) / `1.0` (vLLM) | Controls randomness. Lower (`0.1–0.3`) is deterministic/factual; higher (`0.8+`) is creative. |
| `top_p` | float | `0.9` (Ollama) / `1.0` (vLLM) | Nucleus sampling. Only considers top tokens whose cumulative probability reaches `top_p`. |
| `top_k` | int | `40` (Ollama) / `-1` (vLLM) | Limits token selection to the `k` most probable next tokens (`-1` disables it in vLLM). |
| `min_p` | float | `0.0` | Minimum probability threshold relative to the most likely token (`0.05` is a common setting). |
| `seed` | int | `0` / `null` | Random seed for reproducible outputs. |
| `stop` | string / list | `null` | Sequences that immediately halt generation when encountered. |
| `num_predict` / `max_tokens` | int | `-1` / `null` | Maximum tokens to generate. Ollama uses `-1` for infinite; vLLM uses remaining context. |
| `repeat_penalty` / `repetition_penalty` | float | `1.1` / `1.0` | Penalizes tokens that have already appeared (`> 1.0` discourages repetition). |
| `presence_penalty` | float | `0.0` | Flat penalty applied if a token has appeared at least once (`-2.0` to `2.0`). |
| `frequency_penalty` | float | `0.0` | Penalty proportional to how many times a token has already appeared (`-2.0` to `2.0`). |

---

### 2. Ollama-Only `"options"` Parameters

Ollama passes hardware, context, and llama.cpp-specific controls per request inside `"options"` (or baked into a `Modelfile`):

* **`num_ctx`** *(int, default `2048` or VRAM-scaled)*: Sets the context window size in tokens.
* **`repeat_last_n`** *(int, default `64`)*: How many tokens back the model looks to prevent repetition (`0` = disabled, `-1` = `num_ctx`).
* **`typical_p`** *(float, default `1.0`)*: Locally typical sampling threshold (`1.0` disables it).
* **`tfs_z`** *(float, default `1.0`)*: Tail-free sampling parameter to reduce low-probability tail tokens (`1.0` disables it).
* **`mirostat`** *(int, default `0`)*: Enables Mirostat perplexity control (`0` = disabled, `1` = Mirostat, `2` = Mirostat 2.0).
* **`mirostat_tau`** *(float, default `5.0`)*: Target cross-entropy (coherence vs. diversity balance) when Mirostat is active.
* **`mirostat_eta`** *(float, default `0.1`)*: Learning rate for Mirostat feedback adjustments.
* **`penalize_newline`** *(bool, default `true`)*: Whether to apply repetition penalties to newline characters.
* **`num_gpu`** *(int)*: Number of model layers to offload to GPU (`-1` = all layers, `0` = CPU only).
* **`main_gpu`** *(int, default `0`)*: Primary GPU index when splitting across multiple GPUs.
* **`num_thread`** *(int)*: Number of CPU threads to use during computation.
* **`num_batch`** *(int, default `512`)*: Prompt processing batch size.
* **`num_keep`** *(int)*: Number of initial tokens to retain when context shifts.
* **`use_mmap`** / **`use_mlock`** *(bool)*: Memory-mapping and RAM-locking controls for model weights.

---

### 3. vLLM-Only Request Parameters

Unlike Ollama, vLLM fixes GPU/context settings at server startup (`vllm serve`) via vLLM Engine Arguments, so its per-request API parameters focus on decoding algorithms, logprobs, and constrained outputs:

* **`n`** *(int, default `1`)*: Number of output completions to generate for a single prompt.
* **`best_of`** *(int)*: Generates `best_of` completions server-side and returns the top `n`.
* **`min_tokens`** *(int, default `0`)*: Minimum number of tokens to generate before allowing an EOS/stop token.
* **`ignore_eos`** *(bool, default `false`)*: Ignores end-of-sequence tokens and forces generation up to `max_tokens`.
* **`stop_token_ids`** *(list[int])*: List of specific token IDs that stop generation.
* **`include_stop_str_in_output`** *(bool, default `false`)*: Whether to keep the matched stop string at the end of the returned text.
* **`logprobs`** *(int / bool)*: Returns log probabilities for the top `k` output tokens.
* **`prompt_logprobs`** *(int)*: Returns log probabilities for the prompt tokens.
* **`logit_bias`** *(dict[int, float])*: Adjusts the likelihood of specific token IDs (`-100` to `100`).
* **`guided_json`** *(dict / str)*: Forces output to strictly follow a JSON Schema.
* **`guided_regex`** *(str)*: Forces output to match a regular expression pattern.
* **`guided_choice`** *(list[str])*: Restricts output to one of the exact strings in the list.
* **`guided_grammar`** *(str)*: Forces output to follow a context-free grammar (EBNF).
* **`skip_special_tokens`** *(bool, default `true`)*: Strips special tokens (like `<|eot_id|>`) from the output text.
* **`spaces_between_special_tokens`** *(bool, default `true`)*: Inserts spaces between special tokens during decoding.
