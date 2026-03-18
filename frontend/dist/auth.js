const ADMIN_PASSWORD_KEY = "familytree_admin_password";

function getAdminPassword() {
  return localStorage.getItem(ADMIN_PASSWORD_KEY) || "";
}

function setAdminPassword(password) {
  localStorage.setItem(ADMIN_PASSWORD_KEY, password || "");
}

function clearAdminPassword() {
  localStorage.removeItem(ADMIN_PASSWORD_KEY);
}

function buildAuthHeaders(extra = {}) {
  const headers = { ...extra };
  const pwd = getAdminPassword();
  if (pwd) {
    headers["X-Admin-Password"] = pwd;
  }
  return headers;
}

async function adminFetch(url, options = {}) {
  const opts = { ...options };
  opts.headers = buildAuthHeaders(options.headers || {});
  return fetch(url, opts);
}

async function checkAdminStatus() {
  const resp = await adminFetch("/auth/status");
  if (!resp.ok) {
    throw new Error(await resp.text());
  }
  return resp.json();
}

async function adminLogin(password) {
  const resp = await fetch("/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ password })
  });

  const text = await resp.text();
  if (!resp.ok) {
    throw new Error(text || "登录失败");
  }

  setAdminPassword(password);
  return text ? JSON.parse(text) : { ok: true };
}

async function ensureAdminLogin() {
  const current = getAdminPassword();
  if (current) {
    try {
      const st = await checkAdminStatus();
      if (st && st.authorized) return true;
    } catch (_) {}
  }

  const input = window.prompt("请输入管理员密码");
  if (!input) {
    return false;
  }

  await adminLogin(input.trim());
  return true;
}