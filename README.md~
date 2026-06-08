# Fuzzer GUI

A high-performance, concurrent web fuzzer with a clean, high-density graphical interface built in Go and Fyne. Designed for efficiency and ease of use, this tool allows you to discover web resources quickly and manage your findings in an intuitive, clickable list.

## Features

* **Concurrent Scanning:** Fast, multi-threaded scanning to maximize throughput.
* **Intuitive GUI:** Built with Fyne for a responsive, modern desktop experience.
* **High-Density Results:** Compact UI layout allows you to view more findings at once without sacrificing readability.
* **Flexible Wordlist Selection:** Native file picker integration—no need to hardcode file paths.
* **One-Click Navigation:** Directly open discovered URLs in your default browser from the list.
* **Recursive Options:** Optional depth-based recursion for deep directory discovery.

## Prerequisites

To build and run this project from source, ensure you have:

1.  **Go:** Version 1.20 or later.
2.  **System Dependencies:** Fyne requires specific C libraries to interact with your system's graphics.
    * **Debian/Ubuntu/Kali:**
        ```bash
        sudo apt-get update
        sudo apt-get install libgl1-mesa-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libxxf86vm-dev libxkbcommon-dev
        ```
    * *Windows/macOS: No additional libraries are typically required.*

## Installation & Build

1.  **Clone the repository:**
    ```
    git clone https://github.com/AlexEngleDSU/fuzzer.git
    cd fuzzer
    ```
2.  **Tidy dependencies:**
    ```
    go mod tidy
    ```
3.  **Build the application:**
    ```bash
    make install
    # or
    go build -o fuzzer ./cmd/fuzzer/main.go
    ```
4.  **Run:**
    ```bash
    ./fuzzer
    ```

## Usage

1.  Launch the application.
2.  Enter the target URL in the entry box (e.g., `https://example.com/FUZZ`).
3.  Adjust your recursion settings and depth if needed.
4.  Click **"Select Wordlist"** to load your dictionary file from any location on your machine.
5.  Click **"Start Scan"** and watch the results populate in real-time.
6.  Click any item in the results list to open the discovered URL in your browser.

## Contributing

Contributions are what make the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

## License

Distributed under the MIT License. See `LICENSE` for more information.
