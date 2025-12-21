import AudioRecorder from "./components/AudioRecorder";

function App() {
  return (
    <div className="h-screen">
      <nav className="bg-white px-8 py-6 flex justify-between items-center">
        <h2 className="font-bold text-2xl">SHAZAM</h2>
        <div className="flex items-center gap-6">
          <a
            href="https://github.com/Danztee"
            target="_blank"
            rel="noopener noreferrer"
            className="bg-[#046dff] text-white px-6 py-2.5 rounded-full font-bold text-sm tracking-wide hover:bg-blue-600 transition-colors decoration-0 flex items-center gap-2"
          >
            <svg
              viewBox="0 0 24 24"
              className="w-5 h-5 fill-current"
              xmlns="http://www.w3.org/2000/svg"
            >
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
            </svg>
            GITHUB
          </a>
        </div>
      </nav>

      <div className="bg-[#046dff] h-[calc(100vh-5.5rem)] text-white">
        <div className="grid grid-cols-2 gap-10 max-w-5xl items-center mx-auto h-full">
          <aside className="space-y-2">
            <h1 className="text-6xl font-semibold">
              Identify songs in seconds
            </h1>
            <p className="text-xl">Discover music and more with Shazam</p>

            <p className="text-sm text-gray-300">
              Built from scratch by{" "}
              <a
                href="https://x.com/iamdanztee"
                className="text-white underline"
                rel="noopener noreferrer"
                target="_blank"
              >
                danztee
              </a>
            </p>
          </aside>

          <AudioRecorder />
        </div>
      </div>
    </div>
  );
}

export default App;
