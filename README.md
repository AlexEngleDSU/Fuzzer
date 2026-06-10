# Fuzzer GUI
A high-performance, concurrent web fuzzer with a clean, high-density graphical interface built in Go and Fyne. Designed for efficiency and ease of use, this tool allows you to discover web resources quickly and manage your findings in an intuitive, clickable list.
## Features
* **High-Performance Engine:** Concurrent scanning with a synchronized global ticker for precise request-rate control.
* **Intuitive GUI:** Built with Fyne for a responsive, modern, and high-density interface.
* **Native Integration:** Uses native OS file dialogs for seamless wordlist selection.
* **High-Density Results:** Compact UI layout allows you to view more findings at once without sacrificing readability.
* **One-Click Navigation:** Directly open discovered URLs in your default browser from the list.
* **Recursive Options:** Optional depth-based recursion for deep directory discovery.
##Technical Highlights
* **Synchronized Engine:** Utilizes a global shared-ticker mechanism to ensure predictable request rates across all worker threads.
* **Barrier Synchronization:** Implements sync.WaitGroup to handle batch processing, ensuring clean and reliable transitions between recursion depths.
* **Control:** Built-in Pause/Resume functionality allows for real-time interaction without losing scan state.
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
    ```bash
    git clone https://github.com/AlexEngleDSU/fuzzer.git
    cd fuzzer
    go mod tidy
    make install
    # or
    go build -o fuzzer ./cmd/fuzzer/main.go
    ./fuzzer
    ```
* **Note for Linux Users:** Fyne requires standard desktop development headers. If you encounter build errors, ensure your pkg-config is configured and that you have the required graphics development libraries installed as shown above.
## Usage
1.  Launch the application.
2.  Enter the target URL in the entry box (e.g., `https://example.com/FUZZ`).
3.  Adjust options: Set recurtion, depth, filter status codes, set threads, and set delay
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

## Ethical Use & Legal Disclaimer

This tool is designed for authorized penetration testing and educational purposes only. Using this tool against targets without prior explicit permission from the owner is illegal and unethical. The developer assumes no liability and is not responsible for any misuse or damage caused by this software. Use responsibly.

## License

Distributed under the MIT License. See `LICENSE` for more information.
