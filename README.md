# Website Uptime Monitor

This project is a simple website uptime monitor with a Go backend and an Astro frontend. It periodically checks a list of websites, sends email notifications if a site is down, and displays the status on a web dashboard.

## Project Structure

- `/uptime-monitor`: Contains the Go backend application.
- `/uptime-monitor/frontend`: Contains the Astro frontend application.

## How to Run

You will need two terminals running concurrently to see the full application in action.

### 1. Run the Backend

1.  Navigate to the backend directory:
    ```bash
    cd uptime-monitor
    ```
2.  **Important:** Open `config.json` and replace the placeholder values for `email` with your actual SMTP server details. Otherwise, email notifications will fail.
3.  Run the Go application:
    ```bash
    go run main.go
    ```
The backend server will start, and the API will be available at `http://localhost:8080`.

### 2. Run the Frontend

1.  In a new terminal, navigate to the frontend directory:
    ```bash
    cd uptime-monitor/frontend
    ```
2.  Install the dependencies if you haven't already:
    ```bash
    npm install
    ```
3.  Start the Astro development server:
    ```bash
    npm run dev
    ```
4.  Open your browser and navigate to the URL provided by the Astro dev server (usually `http://localhost:4321`) to see the status dashboard.
