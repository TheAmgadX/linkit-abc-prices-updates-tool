// app.js — LinkIt ABC Price Updater frontend logic
// Communicates with the Go backend via Wails auto-generated bindings:
//   window.go.gui.App.*

'use strict';

// ---- State -------------------------------------------------------------------
const state = {
  outputDir: '',
  imfFilePath: '',
  stockFilePath: '',
};

// ---- DOM references ----------------------------------------------------------
const $ = (id) => document.getElementById(id);

const screens = {
  form:       $('screen-form'),
  processing: $('screen-processing'),
  results:    $('screen-results'),
  error:      $('screen-error'),
};

// ---- Screen navigation -------------------------------------------------------
function showScreen(name) {
  Object.values(screens).forEach((s) => {
    s.classList.remove('active');
    s.classList.remove('centered');
  });
  screens[name].classList.add('active');
  if (name === 'processing' || name === 'error') {
    screens[name].classList.add('centered');
  }
}

// ---- Validation helpers ------------------------------------------------------
function setError(fieldId, msg) {
  const field = $('field-' + fieldId);
  const errEl = $('error-' + fieldId);
  if (!field || !errEl) return;
  errEl.textContent = msg;
  if (msg) {
    field.classList.add('has-error');
    errEl.classList.add('visible');
  } else {
    field.classList.remove('has-error');
    errEl.classList.remove('visible');
  }
}

function clearErrors() {
  ['orgName', 'outputDir', 'imfFile', 'stockFile'].forEach((id) => setError(id, ''));
}

function validateForm() {
  let valid = true;

  const orgName = $('orgName').value.trim();
  if (!orgName) {
    setError('orgName', 'Organisation name is required.');
    valid = false;
  } else {
    setError('orgName', '');
  }

  if (!state.outputDir) {
    setError('outputDir', 'Please select an output directory.');
    valid = false;
  } else {
    setError('outputDir', '');
  }

  if (!state.imfFilePath) {
    setError('imfFile', 'Please select the IMF Excel file.');
    valid = false;
  } else {
    setError('imfFile', '');
  }

  if (!state.stockFilePath) {
    setError('stockFile', 'Please select the Stock Excel file.');
    valid = false;
  } else {
    setError('stockFile', '');
  }

  return valid;
}

// ---- File / directory pickers ------------------------------------------------
async function pickOutputDir() {
  try {
    const path = await window.go.gui.App.SelectOutputDirectory();
    if (path) {
      state.outputDir = path;
      $('outputDir').value = path;
      setError('outputDir', '');
    }
  } catch (e) {
    console.error('SelectOutputDirectory error', e);
  }
}

async function pickIMFFile() {
  try {
    const path = await window.go.gui.App.SelectExcelFile();
    if (path) {
      state.imfFilePath = path;
      $('imfFile').value = path;
      setError('imfFile', '');
    }
  } catch (e) {
    console.error('SelectExcelFile (IMF) error', e);
  }
}

async function pickStockFile() {
  try {
    const path = await window.go.gui.App.SelectExcelFile();
    if (path) {
      state.stockFilePath = path;
      $('stockFile').value = path;
      setError('stockFile', '');
    }
  } catch (e) {
    console.error('SelectExcelFile (Stock) error', e);
  }
}

// ---- Processing --------------------------------------------------------------
async function processFiles() {
  clearErrors();
  if (!validateForm()) return;

  const processBtn = $('btn-process');
  processBtn.disabled = true;
  showScreen('processing');

  const input = {
    orgName:       $('orgName').value.trim(),
    outputDir:     state.outputDir,
    imfFilePath:   state.imfFilePath,
    stockFilePath: state.stockFilePath,
  };

  try {
    const result = await window.go.gui.App.ProcessFiles(input);

    if (result.error) {
      showError(result.error);
    } else {
      showResults(result);
    }
  } catch (e) {
    showError(String(e));
  } finally {
    processBtn.disabled = false;
  }
}

// ---- Results screen ----------------------------------------------------------
function formatNumber(n) {
  return n.toLocaleString();
}

function formatDuration(ms) {
  if (ms < 1000) return ms + 'ms';
  return (ms / 1000).toFixed(1) + 's';
}

function showResults(result) {
  // Stats grid
  const statsGrid = $('stats-grid');
  const stats = [
    { label: 'Products Processed', value: formatNumber(result.productsProcessed), warn: false },
    { label: 'Stock Products',     value: formatNumber(result.stockProducts),     warn: false },
    { label: 'IMF Products',       value: formatNumber(result.imfProducts),       warn: false },
    { label: 'Processing Time',    value: formatDuration(result.durationMs),       warn: false },
    { label: 'Missing from IMF',   value: formatNumber(result.missingInIMF),      warn: result.missingInIMF > 0 },
    { label: 'Missing from Stock', value: formatNumber(result.missingInStock),    warn: result.missingInStock > 0 },
  ];

  statsGrid.innerHTML = stats.map((s) => `
    <div class="stat-card">
      <div class="stat-label">${s.label}</div>
      <div class="stat-value${s.warn ? ' warn' : ''}">${s.value}</div>
    </div>
  `).join('');

  // Reports list
  const reportsList = $('reports-list');
  reportsList.innerHTML = result.outputFiles.map((filename, index) => {
    const isMain = index === 0;
    const iconClass = isMain ? 'ok' : 'warn';
    const icon = isMain ? '✓' : '⚠';
    return `
      <li class="report-item">
        <span class="report-icon ${iconClass}">${icon}</span>
        ${escapeHtml(filename)}
      </li>
    `;
  }).join('');

  // Store output dir for "Open Folder" button
  screens.results.dataset.outputDir = result.outputDir;

  showScreen('results');
}

// ---- Error screen ------------------------------------------------------------
function showError(message) {
  $('error-detail').textContent = message;
  showScreen('error');
}

// ---- Open output folder ------------------------------------------------------
async function openOutputFolder() {
  const dir = screens.results.dataset.outputDir || state.outputDir;
  if (!dir) return;
  try {
    await window.go.gui.App.OpenOutputDirectory(dir);
  } catch (e) {
    console.error('OpenOutputDirectory error', e);
  }
}

// ---- Reset form --------------------------------------------------------------
function resetForm() {
  state.outputDir = '';
  state.imfFilePath = '';
  state.stockFilePath = '';
  $('orgName').value = '';
  $('outputDir').value = '';
  $('imfFile').value = '';
  $('stockFile').value = '';
  clearErrors();
  showScreen('form');
}

// ---- Utility -----------------------------------------------------------------
function escapeHtml(str) {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

// ---- Event wiring ------------------------------------------------------------
document.addEventListener('DOMContentLoaded', () => {
  // File/directory pickers
  $('btn-outputDir').addEventListener('click', pickOutputDir);
  $('btn-imfFile').addEventListener('click', pickIMFFile);
  $('btn-stockFile').addEventListener('click', pickStockFile);

  // Form submission
  $('main-form').addEventListener('submit', (e) => {
    e.preventDefault();
    processFiles();
  });

  // Results actions
  $('btn-open-folder').addEventListener('click', openOutputFolder);
  $('btn-process-another').addEventListener('click', resetForm);

  // Error back button
  $('btn-error-back').addEventListener('click', () => showScreen('form'));

  // Show form on load
  showScreen('form');
});
