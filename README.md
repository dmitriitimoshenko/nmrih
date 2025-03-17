# NMRiH Server Dashboard

This repository hosts a multi-container project that provides a dashboard for monitoring your NMRiH server. The project comprises a backend for log parsing and CSV management (log_api), a responsive, dark-themed React frontend (log_frontend), and a Traefik reverse proxy for secure routing and HTTPS.

## Project Structure

```
nmrih/
├── docker-compose.yaml       # Orchestrates all Docker containers (log_api, log_frontend, traefik, etc.)
├── log_api/                  # Backend API for log parsing and CSV file management
│   ├── Dockerfile            # Dockerfile for building the log_api container
│   └── internal/             # Source code for log parsing, CSV generation, etc.
├── log_frontend/             # React-based dashboard application
│   ├── public/
│   │   └── index.html        # HTML template
│   ├── src/
│   │   ├── components/       # Reusable React components (e.g., TopTimeChart, CountryPieChart, PlayersInfo, Controls)
│   │   ├── hooks/            # Custom hooks (e.g., useTopTimeChartData, useWindowDimensions)
│   │   ├── App.js            # Main application component
│   │   └── App.css           # Global styles (dark theme, responsive design)
│   └── package.json          # Frontend dependencies and scripts
└── traefik/                  # Traefik configuration and SSL certificate storage (acme.json)
```

## Features

- **Backend API (log_api):**
  - Parses server logs and saves data as CSV files.
  - Provides various endpoints for retrieving:
    - Top time-spent players
    - Countries statistics (for the pie chart)
    - Detailed player information (e.g., name, score, formatted session duration)
  
- **Responsive Frontend (log_frontend):**
  - **Top Time-Spent Players:**  
    Displays a bar chart of players sorted by their session durations.
  - **Top Countries:**  
    Shows a pie chart with connection percentages by country plus a legend.
  - **Player Info:**  
    Lists connected players with details such as score and duration, formatted (e.g., `50s`, `50m40s`, `24h59m59s`).
  - **Controls:**  
    Provides buttons for refreshing data and copying the server address to the clipboard (with temporary "Copied!" feedback).
  - Uses a dark theme and adapts its layout based on the device (horizontal on PC, vertical on mobile).

- **Traefik Reverse Proxy:**
  - Provides secure HTTPS support and manages SSL certificates.
  - Routes requests to the appropriate containers (backend or frontend).

- **Dockerized Environment:**
  - Easily build and run the complete project using Docker Compose.

## Installation

### Prerequisites
- Docker and Docker Compose must be installed on your system.

### Steps

1. **Clone the Repository:**
   ```bash
   git clone <repository_url>
   cd nmrih
   ```

2. **Prepare Traefik Certificate Storage:**
   ```bash
   mkdir -p traefik/acme
   touch traefik/acme/acme.json
   chmod 600 traefik/acme/acme.json
   ```
   __Don't forget to change Traefik configurations in `traefik.yaml` and in `docker-compose.yaml` !__

3. **Build and Start Containers:**
   ```bash
   make docker-re-run
   ```

## Usage

- **Access the Dashboard:**  
  Open your browser and navigate to the configured domain (e.g., `https://rulat-bot.duckdns.org`) or use the mapped localhost ports (it should be additionally configured).

- **Dashboard Features:**
  - **Top Time-Spent Players:**  
    View a bar chart that details players’ session durations.
  - **Top Countries:**  
    See a pie chart reflecting connection percentages by country.
  - **Player Info:**  
    Inspect the list of connected players along with their score and session duration.
  - **Online Statistics:**  
    Allows to get insights about what is an average count of concurrent sessions at any hour in a day.
  - **Controls:**  
    Refresh the data or copy the server address using the provided buttons.

## Customization

- **API Endpoints:**  
  Adjust the API URLs in the frontend source code (`log_frontend/src/App.js` and related components) if your backend endpoints or proxy settings change.
- **Styling:**  
  Modify the dark theme and responsive layout via `log_frontend/src/App.css`.
- **Extending Functionality:**  
  Add new components (e.g., additional charts or logs) in the `components/` directory of the frontend.

## Contributing

Contributions, issues, and feature requests are welcome!  
Please check the [Issues](https://github.com/dmitriitimoshenko/nmrih/issues) page or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).

## Contact

For questions or further assistance, please contact via:
- Email: [dmitrii.timoshenko16@gmail.com](mailto:dmitrii.timoshenko16@gmail.com).
- Telegram: [@Kritcz](https://t.me/Kritcz)
