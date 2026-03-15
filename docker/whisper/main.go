package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func main() {
	http.HandleFunc("/inference", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		log.Printf("🎙️ Received transcription request (%d bytes)", r.ContentLength)
		
		tmpWav := fmt.Sprintf("/tmp/voice_%d.wav", time.Now().UnixNano())
		
		// Run ffmpeg to transcode stdin (browser WebM/Opus) to 16kHz Mono WAV file
		// volume=5.0 is a good balance
		ffmpegCmd := exec.Command("ffmpeg", "-i", "pipe:0", "-filter:a", "highpass=f=80,volume=5.0", "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", "-f", "wav", "-y", tmpWav)
		
		ffmpegStdin, _ := ffmpegCmd.StdinPipe()
		var ffmpegErr bytes.Buffer
		ffmpegCmd.Stderr = &ffmpegErr

		if err := ffmpegCmd.Start(); err != nil {
			log.Printf("❌ FFmpeg start error: %v", err)
			http.Error(w, "FFmpeg failed", 500)
			return
		}

		// Pipe r.Body to ffmpeg
		go func() {
			defer ffmpegStdin.Close()
			io.Copy(ffmpegStdin, r.Body)
		}()

		if err := ffmpegCmd.Wait(); err != nil {
			log.Printf("⚠️ FFmpeg error: %v | Stderr: %s", err, ffmpegErr.String())
			http.Error(w, "FFmpeg processing failed", 500)
			os.Remove(tmpWav)
			return
		}

		// Check if wav exists and has size
		fi, err := os.Stat(tmpWav)
		if err != nil || fi.Size() < 100 {
			log.Printf("⚠️ WAV file too small or missing: %v", err)
			http.Error(w, "Audio capture failed", 500)
			return
		}

		// Run whisper on the temp file
		// -nt (no timestamps) for cleaner text, -l auto for multi-language, -t 4 for speed
		whisperCmd := exec.Command("./whisper", "-m", "/models/model.bin", "-f", tmpWav, "-l", "auto", "-nt", "-t", "4")
		
		var outBuf bytes.Buffer
		var whisperErr bytes.Buffer
		whisperCmd.Stdout = &outBuf
		whisperCmd.Stderr = &whisperErr

		if err := whisperCmd.Run(); err != nil {
			log.Printf("⚠️ Whisper error: %v | Stderr: %s", err, whisperErr.String())
			// Don't error out yet, might have partial result
		}

		// Clean up
		os.Remove(tmpWav)

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("X-Equinox-Whisper", "processed")
		w.Write(outBuf.Bytes())
		log.Printf("✅ Transcription complete: %d chars", outBuf.Len())
		if outBuf.Len() > 0 {
			log.Printf("📝 Result: %s", outBuf.String())
		}
	})

	log.Println("🛡️  Whisper-Engine (Zero-Persistence) listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
