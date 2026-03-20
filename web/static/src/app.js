import htmx from "htmx.org";
import { EditorState } from "@codemirror/state";
import { EditorView, keymap, lineNumbers } from "@codemirror/view";
import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
import { markdown } from "@codemirror/lang-markdown";

window.htmx = htmx;

const editors = new Map();

function csrfToken() {
  return document.querySelector('meta[name="csrf-token"]')?.content ?? "";
}

function statusNode() {
  return document.getElementById("file-status");
}

function setStatus(message, kind = "info") {
  const node = statusNode();
  if (!node) return;
  node.textContent = message;
  node.dataset.kind = kind;
}

function editorForTextarea(textarea) {
  return editors.get(textarea);
}

function syncDirty(textarea) {
  const entry = editorForTextarea(textarea);
  if (!entry) return;
  const current = entry.view.state.doc.toString();
  const dirty = current !== entry.initial;
  textarea.dataset.dirty = String(dirty);
  const panel = textarea.closest(".file-panel");
  if (panel) {
    panel.dataset.dirty = String(dirty);
  }
}

function initEditor(textarea) {
  if (textarea.dataset.cmReady === "true") return;
  const readOnly = textarea.dataset.readOnly === "true";
  const host = document.createElement("div");
  host.className = "editor-host";
  textarea.parentNode.insertBefore(host, textarea);
  textarea.hidden = true;

  const state = EditorState.create({
    doc: textarea.value,
    extensions: [
      lineNumbers(),
      history(),
      keymap.of([...defaultKeymap, ...historyKeymap]),
      markdown(),
      EditorView.lineWrapping,
      EditorView.editable.of(!readOnly),
      EditorView.updateListener.of((update) => {
        if (!update.docChanged) return;
        textarea.value = update.state.doc.toString();
        syncDirty(textarea);
      }),
    ],
  });

  const view = new EditorView({ state, parent: host });
  editors.set(textarea, { view, initial: textarea.value });
  textarea.dataset.cmReady = "true";
  syncDirty(textarea);
}

function initEditors(root = document) {
  root.querySelectorAll("textarea.js-editor").forEach(initEditor);
}

function anyDirty() {
  return [...editors.values()].some((entry) => entry.view.state.doc.toString() !== entry.initial);
}

async function saveCurrentFile(button) {
  const panel = button.closest(".file-panel");
  const textarea = panel?.querySelector("textarea.js-editor");
  if (!textarea) return;
  const entry = editorForTextarea(textarea);
  const text = entry ? entry.view.state.doc.toString() : textarea.value;
  button.disabled = true;
  setStatus("Saving…");

  try {
    const response = await fetch(`/api/file?path=${encodeURIComponent(button.dataset.path)}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrfToken(),
      },
      body: JSON.stringify({
        text,
        lastModifiedNs: Number(textarea.dataset.lastModified),
        contentHash: textarea.dataset.contentHash,
      }),
    });

    if (response.status === 409) {
      const payload = await response.json();
      textarea.dataset.lastModified = String(payload.lastModifiedNs);
      textarea.dataset.contentHash = payload.contentHash;
      setStatus(payload.message, "error");
      return;
    }

    if (!response.ok) {
      throw new Error(await response.text());
    }

    const payload = await response.json();
    textarea.dataset.lastModified = String(payload.lastModifiedNs);
    textarea.dataset.contentHash = payload.contentHash;
    const editor = editorForTextarea(textarea);
    if (editor) {
      editor.initial = text;
      syncDirty(textarea);
    }
    const preview = panel.querySelector(".markdown-body");
    if (preview && payload.renderedHtml) {
      preview.innerHTML = payload.renderedHtml;
    }
    setStatus(payload.message, "success");
  } catch (error) {
    setStatus(String(error), "error");
  } finally {
    button.disabled = false;
  }
}

async function copyPath(button) {
  try {
    const response = await fetch("/api/file/copy-path", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-CSRF-Token": csrfToken(),
      },
      body: JSON.stringify({
        path: button.dataset.path,
        kind: button.dataset.kind,
      }),
    });
    if (!response.ok) throw new Error(await response.text());
    const payload = await response.json();
    await navigator.clipboard.writeText(payload.value);
    setStatus(`Copied ${button.dataset.kind.replace("_", " ")}`, "success");
  } catch (error) {
    setStatus(String(error), "error");
  }
}

document.addEventListener("click", (event) => {
  const saveButton = event.target.closest(".js-save-file");
  if (saveButton) {
    event.preventDefault();
    saveCurrentFile(saveButton);
    return;
  }

  const copyButton = event.target.closest(".js-copy-path");
  if (copyButton) {
    event.preventDefault();
    copyPath(copyButton);
  }
});

window.addEventListener("beforeunload", (event) => {
  if (!anyDirty()) return;
  event.preventDefault();
  event.returnValue = "";
});

document.body.addEventListener("htmx:configRequest", (event) => {
  event.detail.headers["X-CSRF-Token"] = csrfToken();
});

document.body.addEventListener("htmx:beforeRequest", (event) => {
  if (!anyDirty()) return;
  const path = event.detail.pathInfo?.requestPath ?? "";
  if (path.startsWith("/api/file?") && event.detail.verb === "get") {
    if (!window.confirm("You have unsaved changes. Continue and lose them?")) {
      event.preventDefault();
    }
  }
});

document.body.addEventListener("htmx:afterSwap", (event) => {
  initEditors(event.target);
});

initEditors();
