package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

const (
	port = 9042
)

type GPUActivity struct {
	GFX         MetricValue `json:"GFX"`
	MediaEngine MetricValue `json:"MediaEngine"`
	Memory      MetricValue `json:"Memory"`
}

type MetricValue struct {
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
}

type VRAM struct {
	TotalVRAM      MetricValue `json:"Total VRAM"`
	TotalVRAMUsage MetricValue `json:"Total VRAM Usage"`
	TotalGTT       MetricValue `json:"Total GTT"`
	TotalGTTUsage  MetricValue `json:"Total GTT Usage"`
}

type Sensors struct {
	AveragePower    MetricValue `json:"Average Power"`
	EdgeTemperature MetricValue `json:"Edge Temperature"`
	JunctionTemp    MetricValue `json:"Junction Temperature"`
}

type GPUMetrics struct {
	AverageSocketPower int     `json:"average_socket_power"`
	CurrentGFXclk      float64 `json:"current_gfxclk"`
	CurrentUclk        float64 `json:"current_uclk"`
	TemperatureEdge    float64 `json:"temperature_edge"`
	TemperatureHotspot float64 `json:"temperature_hotspot"`
}

type DevicePath struct {
	Card   string `json:"card"`
	PCI    string `json:"pci"`
	Render string `json:"render"`
}

type GPUData struct {
	GPUActivity GPUActivity `json:"gpu_activity"`
	VRAM        VRAM        `json:"VRAM"`
	Sensors     Sensors     `json:"Sensors"`
	GPUMetrics  GPUMetrics  `json:"gpu_metrics"`
	DeviceName  string      `json:"DeviceName"`
	ASICName    string      `json:"ASIC Name"`
	DevicePath  DevicePath  `json:"DevicePath"`
}

func fetchGPUMetrics() ([]GPUData, error) {
	cmd := exec.Command("amdgpu_top", "-d", "-gm", "-J")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute amdgpu_top: %v", err)
	}

	var gpuData []GPUData
	if err := json.Unmarshal(output, &gpuData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(gpuData) == 0 {
		return nil, fmt.Errorf("no GPU data found")
	}

	return gpuData, nil
}

func formatPrometheusMetric(name string, value float64, help string, labels ...string) string {
	labelStr := ""
	if len(labels) > 0 {
		labelParts := make([]string, len(labels)/2)
		for i := 0; i < len(labels); i += 2 {
			labelParts[i/2] = fmt.Sprintf(`%s="%s"`, labels[i], labels[i+1])
		}
		labelStr = "{" + join(labelParts, ",") + "}"
	}

	return fmt.Sprintf("# HELP amdgpu_%s %s\n# TYPE amdgpu_%s gauge\namdgpu_%s%s %f\n", name, help, name, name, labelStr, value)
}

func formatPrometheusMetricInt(name string, value int, help string, labels ...string) string {
	return formatPrometheusMetric(name, float64(value), help, labels...)
}

func join(slice []string, sep string) string {
	result := ""
	for i, s := range slice {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	gpuList, err := fetchGPUMetrics()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch GPU metrics: %v", err), http.StatusInternalServerError)
		return
	}

	metrics := make([]string, 0)

	for _, gpu := range gpuList {
		deviceLabels := []string{"device", gpu.DeviceName, "asic", gpu.ASICName, "pci", gpu.DevicePath.PCI}

		// GPU Activity / Compute Load
		if gpu.GPUActivity.GFX.Value >= 0 {
			metrics = append(metrics, formatPrometheusMetric("gpu_usage_percent", gpu.GPUActivity.GFX.Value, "GPU usage percentage (compute/GFX)", deviceLabels...))
		}
		if gpu.GPUActivity.MediaEngine.Value >= 0 {
			metrics = append(metrics, formatPrometheusMetric("media_engine_usage_percent", gpu.GPUActivity.MediaEngine.Value, "Media engine usage percentage", deviceLabels...))
		}
		if gpu.GPUActivity.Memory.Value >= 0 {
			metrics = append(metrics, formatPrometheusMetric("memory_activity_percent", gpu.GPUActivity.Memory.Value, "Memory activity percentage", deviceLabels...))
		}

		// VRAM Usage
		if gpu.VRAM.TotalVRAM.Value > 0 {
			metrics = append(metrics, formatPrometheusMetric("vram_total_mb", gpu.VRAM.TotalVRAM.Value, "Total VRAM in MB", deviceLabels...))
		}
		if gpu.VRAM.TotalVRAMUsage.Value >= 0 {
			metrics = append(metrics, formatPrometheusMetric("vram_used_mb", gpu.VRAM.TotalVRAMUsage.Value, "Used VRAM in MB", deviceLabels...))
		}

		// GTT (VTT) Usage
		if gpu.VRAM.TotalGTT.Value > 0 {
			metrics = append(metrics, formatPrometheusMetric("vtt_total_mb", gpu.VRAM.TotalGTT.Value, "Total GTT/VTT in MB", deviceLabels...))
		}
		if gpu.VRAM.TotalGTTUsage.Value >= 0 {
			metrics = append(metrics, formatPrometheusMetric("vtt_used_mb", gpu.VRAM.TotalGTTUsage.Value, "Used GTT/VTT in MB", deviceLabels...))
		}

		// Power metrics
		if gpu.Sensors.AveragePower.Value > 0 {
			metrics = append(metrics, formatPrometheusMetric("power_usage_watts", gpu.Sensors.AveragePower.Value, "Current power usage in Watts", deviceLabels...))
		}
		if gpu.GPUMetrics.AverageSocketPower > 0 {
			metrics = append(metrics, formatPrometheusMetricInt("socket_power_watts", gpu.GPUMetrics.AverageSocketPower, "Socket power in Watts", deviceLabels...))
		}

		// GPU Frequency
		if gpu.GPUMetrics.CurrentGFXclk > 0 {
			metrics = append(metrics, formatPrometheusMetric("gpu_frequency_mhz", gpu.GPUMetrics.CurrentGFXclk, "Current GPU core frequency in MHz", deviceLabels...))
		}

		// Memory Frequency
		if gpu.GPUMetrics.CurrentUclk > 0 {
			metrics = append(metrics, formatPrometheusMetric("memory_frequency_mhz", gpu.GPUMetrics.CurrentUclk, "Current memory frequency in MHz", deviceLabels...))
		}

		// Temperature metrics
		if gpu.Sensors.EdgeTemperature.Value > 0 {
			metrics = append(metrics, formatPrometheusMetric("edge_temperature_celsius", gpu.Sensors.EdgeTemperature.Value, "GPU edge temperature in Celsius", deviceLabels...))
		}
		if gpu.Sensors.JunctionTemp.Value > 0 {
			metrics = append(metrics, formatPrometheusMetric("junction_temperature_celsius", gpu.Sensors.JunctionTemp.Value, "GPU junction/hotspot temperature in Celsius", deviceLabels...))
		}
		if gpu.GPUMetrics.TemperatureEdge > 0 {
			metrics = append(metrics, formatPrometheusMetric("gpu_temp_edge_celsius", gpu.GPUMetrics.TemperatureEdge, "GPU edge temperature from metrics in Celsius", deviceLabels...))
		}
		if gpu.GPUMetrics.TemperatureHotspot > 0 {
			metrics = append(metrics, formatPrometheusMetric("gpu_temp_hotspot_celsius", gpu.GPUMetrics.TemperatureHotspot, "GPU hotspot temperature from metrics in Celsius", deviceLabels...))
		}

		// Info metric (static)
		metrics = append(metrics, formatPrometheusMetricInt("info", 1, "GPU info metric", deviceLabels...))
	}

	w.Write([]byte(join(metrics, "\n")))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, `{"status":"healthy"}`)
}

func main() {
	http.HandleFunc("/metrics", handleMetrics)
	http.HandleFunc("/health", handleHealth)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting AMD GPU exporter on %s", addr)
	log.Printf("Metrics available at http://localhost%s/metrics", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
