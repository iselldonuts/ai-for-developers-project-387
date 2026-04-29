module.exports = {
  ci: {
    collect: {
      url: ["http://127.0.0.1:8080/", "http://127.0.0.1:8080/owner"],
      numberOfRuns: 3,
      settings: {
        chromeFlags: "--no-sandbox --headless=new",
      },
    },
    upload: {
      target: "filesystem",
      outputDir: "./lighthouse-reports",
      reportFilenamePattern: "%%PATHNAME%%-%%DATETIME%%-report.%%EXTENSION%%",
    },
  },
};
