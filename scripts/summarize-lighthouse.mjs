import { appendFile, mkdir, readFile, readdir, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";

const reportsDir = path.resolve("lighthouse-reports");
const summaryPath = path.join(reportsDir, "summary.md");
const weakAuditLimit = 8;

function formatScore(score) {
  if (typeof score !== "number") {
    return "n/a";
  }

  return `${Math.round(score * 100)}`;
}

function categoryRows(report) {
  return Object.values(report.categories ?? {}).map((category) => {
    return `| ${category.title} | ${formatScore(category.score)} |`;
  });
}

function auditScore(audit) {
  if (typeof audit.score === "number") {
    return audit.score;
  }

  return 1;
}

function isActionableAudit(audit) {
  if (!audit || typeof audit !== "object") {
    return false;
  }

  if (audit.score == null || audit.score >= 0.9) {
    return false;
  }

  return !["notApplicable", "manual", "informative"].includes(audit.scoreDisplayMode);
}

function weakAudits(report) {
  return Object.entries(report.audits ?? {})
    .filter(([, audit]) => isActionableAudit(audit))
    .sort(([, left], [, right]) => auditScore(left) - auditScore(right))
    .slice(0, weakAuditLimit)
    .map(([id, audit]) => {
      const score = formatScore(audit.score);
      const display = audit.displayValue ? ` (${audit.displayValue})` : "";
      return `- ${audit.title}: score ${score}${display} [${id}]`;
    });
}

function pageTitle(report) {
  try {
    const url = new URL(report.finalDisplayedUrl ?? report.finalUrl ?? "");
    return url.pathname === "/" ? "/" : url.pathname;
  } catch {
    return report.finalDisplayedUrl ?? report.finalUrl ?? "unknown URL";
  }
}

async function readReports() {
  const manifestReports = await readManifestReports();
  if (manifestReports.length > 0) {
    return manifestReports.sort((left, right) => pageTitle(left.report).localeCompare(pageTitle(right.report)));
  }

  let entries;

  try {
    entries = await readdir(reportsDir, { withFileTypes: true });
  } catch (error) {
    if (error.code === "ENOENT") {
      return [];
    }

    throw error;
  }

  const reports = [];

  for (const entry of entries) {
    if (!entry.isFile() || !entry.name.endsWith(".json") || entry.name === "manifest.json") {
      continue;
    }

    const filePath = path.join(reportsDir, entry.name);
    const raw = await readFile(filePath, "utf8");
    const report = JSON.parse(raw);

    reports.push({ fileName: entry.name, report });
  }

  return uniqueReports(reports).sort((left, right) => pageTitle(left.report).localeCompare(pageTitle(right.report)));
}

async function readManifestReports() {
  const manifestPath = path.join(reportsDir, "manifest.json");
  let manifest;

  try {
    manifest = JSON.parse(await readFile(manifestPath, "utf8"));
  } catch (error) {
    if (error.code === "ENOENT") {
      return [];
    }

    throw error;
  }

  if (!Array.isArray(manifest)) {
    return [];
  }

  const entries = manifest.filter((entry) => entry.isRepresentativeRun);
  const selectedEntries = entries.length > 0 ? entries : uniqueManifestEntries(manifest);
  const reports = [];

  for (const entry of selectedEntries) {
    if (!entry.jsonPath) {
      continue;
    }

    const filePath = path.isAbsolute(entry.jsonPath) ? entry.jsonPath : path.resolve(entry.jsonPath);
    const raw = await readFile(filePath, "utf8");
    reports.push({
      fileName: path.relative(reportsDir, filePath),
      report: JSON.parse(raw),
    });
  }

  return reports;
}

function uniqueManifestEntries(entries) {
  const byUrl = new Map();

  for (const entry of entries) {
    if (!byUrl.has(entry.url)) {
      byUrl.set(entry.url, entry);
    }
  }

  return [...byUrl.values()];
}

function uniqueReports(reports) {
  const byPage = new Map();

  for (const item of reports) {
    const title = pageTitle(item.report);
    if (!byPage.has(title)) {
      byPage.set(title, item);
    }
  }

  return [...byPage.values()];
}

function buildSummary(reports) {
  const lines = [
    "# Lighthouse report",
    "",
    `Generated: ${new Date().toISOString()}`,
    "",
  ];

  if (reports.length === 0) {
    lines.push("No Lighthouse JSON reports were found.");
    return lines.join("\n");
  }

  lines.push("## Scores", "", "| Page | Performance | Accessibility | Best Practices | SEO |", "| --- | ---: | ---: | ---: | ---: |");

  for (const { report } of reports) {
    const categories = report.categories ?? {};
    lines.push(
      `| ${pageTitle(report)} | ${formatScore(categories.performance?.score)} | ${formatScore(categories.accessibility?.score)} | ${formatScore(categories["best-practices"]?.score)} | ${formatScore(categories.seo?.score)} |`,
    );
  }

  lines.push("", "## Category details", "");

  for (const { fileName, report } of reports) {
    lines.push(`### ${pageTitle(report)}`, "", "| Category | Score |", "| --- | ---: |", ...categoryRows(report), "", `Report JSON: \`${fileName}\``, "");
  }

  lines.push("## Нужны правки", "");

  let hasFixes = false;

  for (const { report } of reports) {
    const items = weakAudits(report);
    if (items.length === 0) {
      continue;
    }

    hasFixes = true;
    lines.push(`### ${pageTitle(report)}`, "", ...items, "");
  }

  if (!hasFixes) {
    lines.push("Критичных или слабых Lighthouse-аудитов с score ниже 90 не найдено.", "");
  }

  return lines.join("\n");
}

const reports = await readReports();
const summary = buildSummary(reports);

await mkdir(reportsDir, { recursive: true });
await writeFile(summaryPath, `${summary}\n`);

if (process.env.GITHUB_STEP_SUMMARY) {
  await appendFile(process.env.GITHUB_STEP_SUMMARY, `${summary}\n`);
}

console.log(`Wrote ${summaryPath}`);
