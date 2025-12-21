import { useState, useRef } from "react";
import MatchResult, { type Match } from "./MatchResult";

const AudioRecorder = () => {
  const [isRecording, setIsRecording] = useState(false);
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [matchResult, setMatchResult] = useState<Match[]>([]);

  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const chunksRef = useRef<Blob[]>([]);
  const timeoutRef = useRef<any>(null);
  const retryCountRef = useRef(0);

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
          } else {
            handleRetryOrFailure();
          }
        };
        reader.readAsDataURL(blob);
        chunksRef.current = [];
      };

      mediaRecorderRef.current.start();
      setIsRecording(true);
      setIsAnalyzing(false);

      timeoutRef.current = setTimeout(() => {
        stopRecording();
      }, 3000);
    } catch (error) {
      console.error("Error accessing microphone:", error);
      alert("Could not access microphone.");
      setIsRecording(false);
      setIsAnalyzing(false);
    }
  };

  const stopRecording = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
    if (
      mediaRecorderRef.current &&
      mediaRecorderRef.current.state === "recording"
    ) {
      mediaRecorderRef.current.stop();
      setIsRecording(false);
      setIsAnalyzing(true);

      mediaRecorderRef.current.stream
        .getTracks()
        .forEach((track) => track.stop());
    }
  };

  const handleRetryOrFailure = () => {
    if (retryCountRef.current < 2) {
      retryCountRef.current += 1;
      startRecording();
    } else {
      retryCountRef.current = 0;
      setIsAnalyzing(false);
      alert("Song not found. Please try again.");
    }
  };

  const sendToBackend = async (base64Data: string) => {
    try {
      const response = await fetch("/api/v1/songs/match", {
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
      const matches = result.records || [];

      if (matches.length > 0) {
        setMatchResult(matches);
        retryCountRef.current = 0;
        setIsAnalyzing(false);
      } else {
        handleRetryOrFailure();
      }
    } catch (error) {
      console.error("Error sending audio to backend:", error);
      handleRetryOrFailure();
    }
  };

  const handleReset = () => {
    setMatchResult([]);
    setIsRecording(false);
    setIsAnalyzing(false);
    retryCountRef.current = 0;
  };

  if (matchResult.length > 0) {
    return <MatchResult match={matchResult[0]} onReset={handleReset} />;
  }

  const isBusy = isRecording || isAnalyzing;

  return (
    <div className="h-screen overflow-hidden bg-[#0474ff] text-white p-4 relative">
      <div className="flex flex-col items-center justify-center h-full gap-8">
        <h3
          className={`text-3xl font-bold transition-all duration-300 ${
            isBusy ? "opacity-0 translate-y-4" : "opacity-100 translate-y-0"
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
          {isAnalyzing && (
            <div className="absolute inset-0 rounded-full border-4 border-white/30 border-t-white animate-spin" />
          )}

          <button
            className={`relative z-10 w-30 h-30 rounded-full border-none bg-white flex items-center justify-center animate-in-out shadow-lg ${
              isBusy
                ? "cursor-not-allowed opacity-80"
                : "cursor-pointer hover:scale-105 transition-transform"
            }`}
            onClick={startRecording}
            disabled={isBusy}
            aria-label="Start Recording"
          >
            <svg className="w-20 h-20 fill-[#0474ff]" viewBox="0 0 24 24">
              <path d="M12 14c1.66 0 3-1.34 3-3V5c0-1.66-1.34-3-3-3S9 3.34 9 5v6c0 1.66 1.34 3 3 3z" />
              <path d="M17 11c0 2.76-2.24 5-5 5s-5-2.24-5-5H5c0 3.53 2.61 6.43 6 6.92V21h2v-3.08c3.39-.49 6-3.39 6-6.92h-2z" />
            </svg>
          </button>
        </div>

        <h3
          className={`text-base font-semibold transition-all duration-300 ${
            isBusy ? "opacity-100 translate-y-0" : "opacity-0 -translate-y-4"
          }`}
        >
          {isAnalyzing ? "Looking for a match..." : "Listening for music..."}
        </h3>
      </div>
    </div>
  );
};

export default AudioRecorder;
