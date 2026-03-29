import { createHash, createPrivateKey, createPublicKey, randomBytes, sign } from "node:crypto";
import { readFile } from "node:fs/promises";
import http from "node:http";
import https from "node:https";

function listen(server, port) {
  return new Promise((resolve, reject) => {
    server.once("error", reject);
    server.listen(port, "127.0.0.1", () => {
      server.off("error", reject);
      resolve();
    });
  });
}

function close(server) {
  return new Promise((resolve, reject) => {
    server.close((error) => {
      if (error) {
        reject(error);
        return;
      }
      resolve();
    });
  });
}

function parseRequestBody(request) {
  return new Promise((resolve, reject) => {
    let body = "";

    request.setEncoding("utf8");
    request.on("data", (chunk) => {
      body += chunk;
    });
    request.on("end", () => resolve(body));
    request.on("error", reject);
  });
}

function sendJSON(response, statusCode, payload) {
  const body = JSON.stringify(payload);
  response.writeHead(statusCode, {
    "content-type": "application/json",
    "content-length": Buffer.byteLength(body),
    "cache-control": "no-store",
  });
  response.end(body);
}

function sendHTML(response, statusCode, markup) {
  response.writeHead(statusCode, {
    "content-type": "text/html; charset=utf-8",
    "content-length": Buffer.byteLength(markup),
    "cache-control": "no-store",
  });
  response.end(markup);
}

function sendNotFound(response) {
  response.writeHead(404, { "content-type": "text/plain; charset=utf-8" });
  response.end("not found");
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll("\"", "&quot;")
    .replaceAll("'", "&#39;");
}

function isSafeOIDCTransportValue(value) {
  return /^[A-Za-z0-9._~-]{1,512}$/.test(String(value ?? ""));
}

function normalizeAllowedLogoutRedirect(rawValue, allowedOrigin) {
  const trimmed = String(rawValue ?? "").trim();
  if (!trimmed) {
    return "";
  }

  try {
    const parsed = new URL(trimmed);
    if (parsed.origin !== allowedOrigin) {
      return "";
    }
    if (parsed.username || parsed.password || parsed.hash) {
      return "";
    }
    return parsed.toString();
  } catch {
    return "";
  }
}

function codeChallengeS256(verifier) {
  return createHash("sha256").update(String(verifier)).digest("base64url");
}

function signJWT(privateKey, header, payload) {
  const encodedHeader = Buffer.from(JSON.stringify(header)).toString("base64url");
  const encodedPayload = Buffer.from(JSON.stringify(payload)).toString("base64url");
  const signingInput = `${encodedHeader}.${encodedPayload}`;
  const signature = sign("RSA-SHA256", Buffer.from(signingInput), privateKey).toString("base64url");
  return `${signingInput}.${signature}`;
}

export async function startLocalHTTPSProxy({ certPath, keyPath, listenPort, targetPort }) {
  const [cert, key] = await Promise.all([
    readFile(certPath, "utf8"),
    readFile(keyPath, "utf8"),
  ]);

  const server = https.createServer({ cert, key }, (request, response) => {
    const upstream = http.request(
      {
        hostname: "127.0.0.1",
        port: targetPort,
        method: request.method,
        path: request.url,
        headers: {
          ...request.headers,
          host: `127.0.0.1:${targetPort}`,
          "x-forwarded-proto": "https",
          "x-forwarded-host": `127.0.0.1:${listenPort}`,
        },
      },
      (upstreamResponse) => {
        const headers = { ...upstreamResponse.headers };
        response.writeHead(upstreamResponse.statusCode ?? 502, headers);
        upstreamResponse.pipe(response);
      },
    );

    upstream.once("error", () => {
      if (!response.headersSent) {
        response.writeHead(502, { "content-type": "text/plain; charset=utf-8" });
      }
      response.end("upstream unavailable");
    });

    request.pipe(upstream);
  });

  await listen(server, listenPort);
  return {
    port: listenPort,
    close: async () => close(server),
  };
}

export async function startLocalOIDCProvider({
  certPath,
  keyPath,
  listenPort,
  clientID,
  clientSecret,
  redirectURL,
  issuerURL,
  testEmail,
  testSubject,
  testName,
  emailVerified,
}) {
  const [cert, key] = await Promise.all([
    readFile(certPath, "utf8"),
    readFile(keyPath, "utf8"),
  ]);
  const privateKey = createPrivateKey(key);
  const publicJWK = createPublicKey(key).export({ format: "jwk" });
  const keyID = "ovumcy-e2e-oidc";
  const authCodes = new Map();
  const redirectOrigin = new URL(redirectURL).origin;

  const jwksPayload = {
    keys: [
      {
        ...publicJWK,
        alg: "RS256",
        kid: keyID,
        use: "sig",
      },
    ],
  };

  const metadata = {
    issuer: issuerURL,
    authorization_endpoint: `${issuerURL}/authorize`,
    token_endpoint: `${issuerURL}/token`,
    jwks_uri: `${issuerURL}/keys`,
    end_session_endpoint: `${issuerURL}/logout`,
    response_types_supported: ["code"],
    subject_types_supported: ["public"],
    id_token_signing_alg_values_supported: ["RS256"],
    token_endpoint_auth_methods_supported: ["client_secret_post", "client_secret_basic"],
  };

  const server = https.createServer({ cert, key }, async (request, response) => {
    try {
      const url = new URL(request.url ?? "/", issuerURL);

      if (request.method === "GET" && url.pathname === "/.well-known/openid-configuration") {
        sendJSON(response, 200, metadata);
        return;
      }

      if (request.method === "GET" && url.pathname === "/keys") {
        sendJSON(response, 200, jwksPayload);
        return;
      }

      if (request.method === "GET" && url.pathname === "/authorize") {
        const redirectURI = url.searchParams.get("redirect_uri") ?? "";
        const state = url.searchParams.get("state") ?? "";
        const nonce = url.searchParams.get("nonce") ?? "";
        const requestedClientID = url.searchParams.get("client_id") ?? "";
        const codeChallenge = url.searchParams.get("code_challenge") ?? "";
        const codeChallengeMethod = url.searchParams.get("code_challenge_method") ?? "";
        const responseMode = url.searchParams.get("response_mode") ?? "";

        if (
          requestedClientID !== clientID ||
          redirectURI !== redirectURL ||
          state === "" ||
          nonce === "" ||
          !isSafeOIDCTransportValue(state) ||
          !isSafeOIDCTransportValue(nonce) ||
          codeChallenge === "" ||
          codeChallengeMethod !== "S256" ||
          responseMode !== "form_post"
        ) {
          sendJSON(response, 400, { error: "invalid_request" });
          return;
        }

        const code = randomBytes(24).toString("hex");
        authCodes.set(code, {
          nonce,
          redirectURI,
          codeChallenge,
          email: testEmail,
          sub: testSubject,
          name: testName,
          emailVerified,
        });

        const callbackMarkup = `<!doctype html>
<html lang="en">
  <body>
    <form id="oidc-callback-form" action="${escapeHTML(redirectURL)}" method="post">
      <input type="hidden" name="code" value="${code}">
      <input type="hidden" name="state" value="${escapeHTML(state)}">
    </form>
    <script>document.getElementById("oidc-callback-form").submit();</script>
  </body>
</html>`;
        sendHTML(response, 200, callbackMarkup);
        return;
      }

      if (request.method === "POST" && url.pathname === "/token") {
        const body = new URLSearchParams(await parseRequestBody(request));
        const code = body.get("code") ?? "";
        const verifier = body.get("code_verifier") ?? "";
        const redirectURI = body.get("redirect_uri") ?? "";
        let requestedClientID = body.get("client_id") ?? "";
        let requestedClientSecret = body.get("client_secret") ?? "";
        const authorization = String(request.headers.authorization ?? "");
        if ((!requestedClientID || !requestedClientSecret) && authorization.startsWith("Basic ")) {
          const decoded = Buffer.from(authorization.slice("Basic ".length), "base64").toString(
            "utf8",
          );
          const separator = decoded.indexOf(":");
          if (separator >= 0) {
            requestedClientID = decoded.slice(0, separator);
            requestedClientSecret = decoded.slice(separator + 1);
          }
        }

        const record = authCodes.get(code);
        if (
          !record ||
          requestedClientID !== clientID ||
          requestedClientSecret !== clientSecret ||
          redirectURI !== record.redirectURI ||
          codeChallengeS256(verifier) !== record.codeChallenge
        ) {
          sendJSON(response, 400, { error: "invalid_grant" });
          return;
        }

        authCodes.delete(code);
        const issuedAt = Math.floor(Date.now() / 1000);
        const idToken = signJWT(
          privateKey,
          { alg: "RS256", typ: "JWT", kid: keyID },
          {
            iss: issuerURL,
            sub: record.sub,
            aud: clientID,
            exp: issuedAt + 300,
            iat: issuedAt,
            nonce: record.nonce,
            email: record.email,
            email_verified: record.emailVerified,
            name: record.name,
          },
        );

        sendJSON(response, 200, {
          access_token: randomBytes(24).toString("hex"),
          token_type: "Bearer",
          expires_in: 300,
          id_token: idToken,
        });
        return;
      }

      if ((request.method === "GET" || request.method === "POST") && url.pathname === "/logout") {
        const redirectURI = normalizeAllowedLogoutRedirect(
          url.searchParams.get("post_logout_redirect_uri"),
          redirectOrigin,
        );
        if (redirectURI) {
          response.writeHead(303, { location: redirectURI, "cache-control": "no-store" });
          response.end();
          return;
        }
        if (url.searchParams.has("post_logout_redirect_uri")) {
          sendJSON(response, 400, { error: "invalid_request" });
          return;
        }
        sendHTML(response, 200, "<!doctype html><html lang=\"en\"><body>signed out</body></html>");
        return;
      }

      sendNotFound(response);
    } catch (error) {
      if (error instanceof Error) {
        console.error("[e2e-runtime] local OIDC provider error:", error.message);
      } else {
        console.error("[e2e-runtime] local OIDC provider error");
      }
      sendJSON(response, 500, { error: "server_error" });
    }
  });

  await listen(server, listenPort);
  return {
    issuerURL,
    close: async () => close(server),
  };
}
