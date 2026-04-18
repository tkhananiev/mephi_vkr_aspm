package agentdebug

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

// Log пишет NDJSON в stderr (docker logs) и при возможности в файл сессии отладки.
func Log(hypothesisID, location, message string, data map[string]any) {
	payload := map[string]any{
		"sessionId":    "cf16da",
		"hypothesisId": hypothesisID,
		"location":     location,
		"message":      message,
		"data":         data,
		"timestamp":    time.Now().UnixMilli(),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	log.Printf("[debug-cf16da] %s", string(b))
	paths := []string{
		"/Users/tvkhananiev/Documents/MEPHI/ВКР/vkr_temp_arch/.cursor/debug-cf16da.log",
		"/opt/mephi_vkr_aspm/.cursor/debug-cf16da.log",
	}
	for _, p := range paths {
		if f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			_, _ = f.Write(append(b, '\n'))
			_ = f.Close()
			break
		}
	}
}
