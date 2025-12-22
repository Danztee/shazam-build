# shazam-build

> A from-scratch implementation of the Shazam audio fingerprinting algorithm in Go.

This project recreates the core logic behind Shazam, as described in the original 2003 paper _"An Industrial-Strength Audio Search Algorithm" by Avery Li-Chun Wang_. It features a **Digital Signal Processing (DSP) pipeline**, a custom **PostgreSQL-based fingerprint database**, and a **React frontend** for real-time audio recognition.

## 🎵 How It Works

This system does not use machine learning or external fingerprinting libraries. Instead, it relies on pure signal processing and probabilistic hashing:

1.  **Spectrogram Generation:** Converts raw audio (WAV) into a time-frequency spectrogram using Fast Fourier Transform (FFT).
2.  **Constellation Map:** Identifies high-energy peaks (local maxima) in the spectrogram to create a sparse representation of the audio.
3.  **Combinatorial Hashing:** Generates unique hashes by pairing "anchor" peaks with nearby "target" peaks and their time deltas. This makes the fingerprints valid even in noisy environments.
4.  **Matching & Time Coherency:** Matches query fingerprints against the database and uses **diagonal alignment** (linearity search) to filter out false positives. If the time offsets of the matching hashes align, it's a match.

## 🛠️ Tech Stack

### Backend

- **Go 1.25+**: Core logic, high-performance DSP pipeline.
- **Chi**: Lightweight router for the REST API.
- **PostgreSQL**: Relational database for storing song metadata and millions of fingerprints.
- **pgx**: High-performance PostgreSQL driver.
- **FFmpeg**: Used for normalizing audio input (converting various formats to 44.1kHz mono WAV).

### Frontend

- **React 19**: Modern UI library.
- **Vite**: Fast build tool.
- **TailwindCSS v4**: Styling.
- **MediaRecorder API**: Capturing microphone input in the browser.

## 🚀 Getting Started

### Prerequisites

Ensure you have the following installed on your system:

- **Go** (v1.25 or later)
- **Node.js** & **pnpm**
- **Docker** & **Docker Compose**
- **FFmpeg** (must be available in your system path)
- **goose** (for database migrations: `go install github.com/pressly/goose/v3/cmd/goose@latest`)
- **sqlc** (optional, for regenerating DB queries)

### Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/Danztee/shazam-build.git
    cd shazam-build
    ```

2.  **Environment Setup:**
    Create a `.env` file in the root directory. You can copy the example:

    ```bash
    cp .env.example .env
    ```

    Ensure your `.env` contains the necessary variables (e.g., `DATABASE_URL`, `PORT`, `SPOTIFY_CLIENT_ID`, `SPOTIFY_CLIENT_SECRET` for ingestion).

3.  **Start the Database:**
    Use the Makefile to spin up the PostgreSQL container:

    ```bash
    make docker-run
    ```

4.  **Run Migrations:**
    Apply the database schema to your local Postgres instance:
    ```bash
    make migrate-up
    ```

### Running the Application

You can start both the backend and frontend with a single command:

```bash
make run
```

- **Backend API:** `http://localhost:8080`
- **Frontend:** `http://localhost:5173`

## 📖 Usage

### Adding Songs (Ingestion)

The system needs a reference database of songs to recognize them. You can add songs via the API:

**Endpoint:** `POST /api/v1/songs`

**Payload:**

```json
{
  "spotify_url": "https://open.spotify.com/track/..."
}
```

_This will fetch the metadata, download the audio, process it through the DSP pipeline, and store the fingerprints._

### Recognizing Songs

1.  Open the frontend at `http://localhost:5173`.
2.  Grant microphone permissions.
3.  Click the **Listen** (or Shazam-like) button.
4.  Play a song in the background.
5.  After a few seconds of recording, the app will send the audio clip to the backend and display the matched result.

## 📂 Project Structure

```
├── cmd/                # Entry points (API server)
├── frontend/           # React + Vite application
├── internal/
│   ├── audio/          # DSP logic (FFT, Spectrogram, etc.)
│   ├── database/       # Database connections and queries
│   ├── songs/          # Business logic for Song management & Matching
│   ├── spotify/        # Spotify API integration
│   └── ...
├── migrations/         # SQL migrations (Goose)
├── queries/            # SQLC queries
├── LINKEDIN_POST.md    # Detailed technical write-up
└── Makefile            # Command shortcuts
```

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

## 👤 Author

**Danztee**
