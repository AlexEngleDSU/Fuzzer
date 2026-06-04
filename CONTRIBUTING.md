# Contributing to Fuzzer

Thank you for your interest in contributing to the Fuzzer project! We welcome community contributions to help make this tool more effective and robust.

---

## How to Contribute

### 1. Reporting Issues
If you encounter a bug or have a feature request:
* Search for existing issues: Check the Issue Tracker to ensure the problem hasn't already been reported.
* Create a new issue: If you find a bug, please provide:
    * A clear description of the issue.
    * Steps to reproduce the error.
    * Your environment details (OS, Go version, Fyne version).

### 2. Submitting Pull Requests
We encourage pull requests (PRs) for bug fixes and new features. To ensure a smooth process, please follow these steps:

1. Fork the Repository: Click the "Fork" button on the top right of the GitHub page.
2. Create a Branch: Create a new branch for your feature or fix.
   git checkout -b feature/your-feature-name
3. Make Changes: Write your code, following the existing style and project structure.
4. Test: Ensure your changes compile and perform as expected.
   go build -o fuzzer ./cmd/fuzzer/main.go
   ./fuzzer
5. Commit and Push:
   git add .
   git commit -m "Add feature/fix: [Description of your changes]"
   git push origin feature/your-feature-name
6. Open a Pull Request: Navigate to the main repository on GitHub and click "Compare & pull request."

---

## Development Guidelines

* Structure:
    * cmd/: Contains the entry point for the application.
    * pkg/: Contains the core logic and GUI components.
* Dependencies: We use Fyne for the GUI. Please ensure you have the necessary development headers installed for your Linux distribution (e.g., libgl1-mesa-dev, libx11-dev).
* Formatting: Please maintain consistent formatting. Running 'go fmt ./...' before committing is highly recommended.

---

## Code of Conduct
By participating in this project, you are expected to uphold our commitment to an open, friendly, and welcoming community for all developers. Please be respectful in your code reviews and comments.

---

## Need Help?
If you have questions about the codebase or the development process, feel free to open an issue or reach out through the project's communication channels.

Happy Coding!
