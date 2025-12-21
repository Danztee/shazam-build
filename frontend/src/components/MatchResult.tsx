import { useEffect, useState } from "react";

export type Match = {
  id: number;
  title: string;
  artists: string[];
  album: string;
  image_url: string;
  duration_ms: number;
  created_at: string;
};

interface MatchResultProps {
  match: Match;
  onReset: () => void;
}

const MatchResult = ({ match, onReset }: MatchResultProps) => {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    setIsVisible(true);
  }, []);

  return (
    <div
      className={`overflow-hidden absolute inset-0 z-50 flex flex-col items-center justify-between bg-black/90 text-white p-6 transition-opacity duration-1000 ${
        isVisible ? "opacity-100" : "opacity-0"
      }`}
    >
      <div
        className="absolute inset-0 bg-cover bg-center opacity-40 blur-3xl scale-125 z-0"
        style={{ backgroundImage: `url(${match.image_url})` }}
      />
      <div className="absolute inset-0 bg-linear-to-b from-black/30 via-transparent to-black/90 z-0" />

      <div className="relative z-10 w-full flex justify-between items-center pt-4">
        <button
          onClick={onReset}
          className="p-2 rounded-full bg-white/10 backdrop-blur-md hover:bg-white/20 transition-all"
        >
          <svg
            className="w-6 h-6 text-white"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M15 19l-7-7 7-7"
            />
          </svg>
        </button>
        <div className="w-10" />
      </div>

      <div className="relative z-10 flex flex-col items-center w-full max-w-sm space-y-8 my-auto">
        <div
          className={`relative w-64 h-64 rounded-2xl overflow-hidden shadow-2xl shadow-blue-500/20 transition-all duration-1000 delay-300 ${
            isVisible ? "translate-y-0 opacity-100" : "translate-y-10 opacity-0"
          }`}
        >
          <img
            src={match.image_url}
            alt={match.album}
            className="w-full h-full object-cover"
          />
        </div>

        <div
          className={`text-center space-y-2 transition-all duration-1000 delay-500 ${
            isVisible ? "translate-y-0 opacity-100" : "translate-y-10 opacity-0"
          }`}
        >
          <h1 className="text-3xl font-bold leading-tight line-clamp-2">
            {match.title}
          </h1>
          <p className="text-xl text-blue-400 font-medium">
            {match.artists.join(", ")}
          </p>
          <p className="text-sm text-gray-400 uppercase tracking-wider mt-1">
            {match.album}
          </p>
        </div>
      </div>

      <div
        className={`relative z-10 w-full max-w-sm space-y-4 pb-8 transition-all duration-1000 delay-700 ${
          isVisible ? "translate-y-0 opacity-100" : "translate-y-10 opacity-0"
        }`}
      >
        <button className="w-full py-4 bg-[#1DB954] hover:bg-[#1ed760] text-black font-bold rounded-full transition-transform active:scale-95 flex items-center justify-center gap-2 shadow-lg shadow-green-500/20 cursor-pointer">
          <svg className="w-6 h-6" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141 4.2-1.32 9.6-0.66 13.261 1.56.479.18.659.66.36.96-.06.241-.18.361-.18.361zM19.2 8.1c-3.839-2.28-10.139-2.46-13.799-1.32-.6.181-1.2-.179-1.38-.721-.18-.6.12-1.2.72-1.38C8.82 3.48 15.839 3.66 20.219 6.24c.6.36.841 1.08.481 1.68-.362.6-1.08.84-1.68.48h.18z" />
          </svg>
          Open in Spotify
        </button>

        {/* <button
          onClick={onReset}
          className="w-full py-4 bg-white/10 hover:bg-white/20 backdrop-blur-md text-white font-semibold rounded-full transition-transform active:scale-95 cursor-pointer border border-white/10"
        >
          Identify Another Song
        </button> */}
      </div>
    </div>
  );
};

export default MatchResult;
