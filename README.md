# amdgpu_exporter

Provides Prometheus-compatible metrics about the AMD GPU on `:9042/metrics`.
Requires `amdgpu_top` installed.

## (!) This project was fully vibe-coded

I wanted to push this model and see what happens.

* Model: [Qwen 3.5 27B](https://huggingface.co/bartowski/Qwen_Qwen3.5-27B-GGUF) in IQ4_XS quant.
* Running on: Llama.cpp `llama-server -v --parallel 1 -hf bartowski/Qwen_Qwen3.5-27B-GGUF:IQ4_XS --jinja --temp 0.7 --min-p 0.01 --top-p 0.80 --top-k 20 --repeat-penalty 1.05 --ctx-size 80000 --host 0.0.0.0 --port 8012 --metrics -ngl auto -fa on -ctk q8_0 -ctv q8_0`
* Tool: [OpenCode](https://opencode.ai/)
* Prompt:

> Write me a service that exports AMD GPU metrics including all the power usage, compute load and memory usage metrics on a prometheus-compatible /metrics endpoint. The command line tool `amdgpu_top` provides all interesting data: `amdgpu_top -gm` provides lots of metrics, `amdgpu_top -d` provides memory usage and further information. There's a "-J" option that outputs all the data in a parseable JSON format. All combined (`amdgpu_top -d -gm -J`) should provide all data needed.
> Have the service run amdgpu_top to accquire the metrics, and have it serve the metrics on HTTP port 9042 under the /metrics path in prometheus compatible format.
> Choose an appropriate programming language and associated frameworks/libraries/packages appropiate for this task (Just please avoid NodeJS or JavaScript/TypeScript). Implement the service. Assume you can run the `amdgpu_top` command with the proposed options to study the output for creating the parser, there is a local AMD GPU running so it should provide realistic real-world data. Some examples for testing: GPU usage should be reported in percent (0-100). GPU frequency in MHz for the present card is somewhere between 1 and 2550MHz. VRAM usage should be betwen 0 and 16384MB. VTT usage somewhere between 0 and 8192MB.
> On regular checkpoints/milestones, create a git commit to save progress.

(Apart from this README.md and LICENSE)
