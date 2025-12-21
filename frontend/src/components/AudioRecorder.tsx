import { useState, useRef } from "react";
const AudioRecorder = () => {
  const [isRecording, setIsRecording] = useState(false);
  const [matchResult, setMatchResult] = useState<any>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);

  const startRecording = async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      mediaRecorderRef.current = new MediaRecorder(stream);

      mediaRecorderRef.current.ondataavailable = (e) => {
        if (e.data.size > 0) {
          chunksRef.current.push(e.data);
        }
      };

      mediaRecorderRef.current.onstop = () => {
        const blob = new Blob(chunksRef.current, { type: "audio/webm" });
        const reader = new FileReader();
        reader.onloadend = async () => {
          const base64 = reader.result?.toString().split(",")[1];
          if (base64) {
            await sendToBackend(base64);
          }
        };
        reader.readAsDataURL(blob);
        chunksRef.current = [];
      };

      mediaRecorderRef.current.start();
      setIsRecording(true);
    } catch (error) {
      console.error("Error accessing microphone:", error);
      alert(
        "Could not access microphone. Please ensure permissions are granted."
      );
    }
  };

  const stopRecording = () => {
    if (mediaRecorderRef.current && isRecording) {
      mediaRecorderRef.current.stop();
      setIsRecording(false);
      // Stop all tracks to release the microphone
      mediaRecorderRef.current.stream
        .getTracks()
        .forEach((track) => track.stop());
    }
  };

  const sendToBackend = async (base64Data: string) => {
    try {
      const response = await fetch("/api/songs/match", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ data: base64Data }),
      });

      if (!response.ok) {
        throw new Error("Failed to match audio");
      }

      const result = await response.json();
      setMatchResult(result);
      console.log("Match result:", result);
    } catch (error) {
      console.error("Error sending audio to backend:", error);
      alert("Failed to match audio. Please try again.");
    }
  };

  return (
    <div className="h-screen overflow-hidden bg-[#0474ff] text-white p-4">
      <div className="flex flex-col items-center justify-center h-full gap-8">
        <h3
          className={`text-3xl font-bold transition-all duration-300 ${
            isRecording
              ? "opacity-0 translate-y-4"
              : "opacity-100 translate-y-0"
          }`}
        >
          Tap to shazam
        </h3>

        <div className="relative flex items-center justify-center">
          {isRecording && (
            <>
              <div className="absolute inset-0 rounded-full bg-white/30 animate-ripple" />
              <div className="absolute inset-0 rounded-full bg-white/30 animate-ripple [animation-delay:0.8s]" />
              <div className="absolute inset-0 rounded-full bg-white/30 animate-ripple [animation-delay:1.6s]" />
            </>
          )}
          <button
            className="cursor-pointer relative z-10 w-30 h-30 rounded-full border-none bg-white flex items-center justify-center animate-in-out shadow-lg"
            onClick={isRecording ? stopRecording : startRecording}
            aria-label={isRecording ? "Stop Recording" : "Start Recording"}
          >
            <svg className="w-20 h-20 fill-[#0474ff]" viewBox="0 0 24 24">
              <path d="M12 14c1.66 0 3-1.34 3-3V5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3z" />
              <path d="M17 11c0 2.76-2.24 5-5 5s-5-2.24-5-5H5c0 3.53 2.61 6.43 6 6.92V21h2v-3.08c3.39-.49 6-3.39 6-6.92h-2z" />
            </svg>
          </button>
        </div>

        <h3
          className={`text-base font-semibold transition-all duration-300 ${
            isRecording
              ? "opacity-100 translate-y-0"
              : "opacity-0 -translate-y-4"
          }`}
        >
          Listening for music...
        </h3>
      </div>
    </div>
  );
};

export default AudioRecorder;
