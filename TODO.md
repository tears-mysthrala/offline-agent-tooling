# TODO: Herramientas locales para Agentes (rápidas, sin dependencias externas y con modo offline)

Este documento define el backlog y especificaciones detalladas para crear herramientas, módulos y funciones que los agentes puedan usar localmente, con foco en:

- Performance y latencia bajas
- Cero o mínimas dependencias externas
- Funcionamiento sin Internet (modo offline) para entornos aislados
- Portabilidad (Windows-first, pero extensible a Linux/macOS)
- Interfaz uniforme de uso vía CLI y API (stdin/stdout JSON) para orquestación por agentes

Nota de entorno: Este repo usa un wrapper de comandos. Para ejecutar colas de comandos:

1) Añade líneas en `scripts/commands.txt`
2) Ejecuta: `pwsh -File scripts/command-runner.ps1`

## Estado de Implementación (Actualizado)

✅ **MVP Complete (7/7)**: FS Tool, Process Runner, Search, Log, HTTP/MOCK, Config, Tool Adapter  
✅ **Plus Complete (5/5)**: Cache, Git, Template, Archive, KV Store  
📝 **Pendiente**: Herramientas avanzadas (Scheduler, AST Parser, Diff/Patch, Formatter, Secrets)

**Documentación**: Ver [README.md](./README.md) y [USAGE.md](./USAGE.md) para ejemplos de uso.


## Principios de diseño

- Local-first: todo debe funcionar sin red; si hay funcionalidades que normalmente usan red, proveer mocks/fixtures y banderas `--offline`.
- Estándares del sistema: priorizar PowerShell (Windows) y Python (stdlib) para evitar dependencias. Node.js sólo si aporta valor claro y con fallback.
- Interfaz consistente: cada herramienta expone
  - CLI: `tool <command> [--flags]`
  - API JSONL: entradas por stdin (línea JSON) y salidas por stdout (línea JSON)
- Observabilidad: logs estructurados JSON en `logs/` y códigos de salida claros.
- Seguridad por defecto: no filtrar secretos en logs, respetar permisos, y validar entradas.
- Idempotencia cuando sea posible; dry-run y `--confirm` para operaciones destructivas.

## Estructura propuesta del proyecto

- `tools/`
  - `ps/` (PowerShell puro, sin módulos externos)
  - `py/` (Python 3.x, sólo stdlib)
  - `shared/` (esquemas JSON, plantillas, fixtures)
- `fixtures/` (datos de prueba y mocks para modo offline)
  - `http/` (respuestas pregrabadas por ruta y query)
  - `git/` (repos de juguete para test)
  - `search/` (árboles de prueba)
- `docs/` (contratos de herramientas, ejemplos)

Nota: En esta fase sólo definimos el backlog; la estructura se creará en entregas sucesivas.

## Contrato estándar de herramientas (API/CLI)

- Entrada JSON por stdin (una línea por operación). Claves comunes:
  - `op`: operación/acción a realizar (string)
  - `args`: parámetros de la operación (objeto)
  - `offline`: booleano (forzar modo sin red)
  - `timeout_ms`: entero opcional
  - `trace_id`: correlación opcional
- Salida JSON por stdout:
  - `ok`: booleano
  - `data`: resultado específico
  - `error`: objeto con `code`, `message`, `details` si `ok=false`
- CLI: banderas equivalentes (`--op`, `--offline`, `--timeout-ms`, etc.).

## Herramientas planificadas (especificaciones)

A continuación se listan las herramientas con: Objetivo, Idiomas, Dependencias, Contrato (inputs/outputs), Requisitos, Casos límite, Offline, Pruebas mínimas, Criterios de aceptación.

---

### 1) FS Tool (sistema de archivos)

- Objetivo: Operaciones básicas y seguras de archivos/carpetas: leer, escribir, listar, copiar, mover, borrar, mkdir -p, glob, lectura/escritura JSON/YAML, checksum.
- Idiomas: PowerShell (prioridad), Python como alternativa.
- Dependencias: stdlib (PowerShell/.NET; Python: `os`, `pathlib`, `glob`, `json`, `hashlib`).
- Contrato:
  - op: `read|write|append|copy|move|delete|mkdir|list|glob|stat|checksum|read-json|write-json|read-lines|write-lines`
  - args: `{ path, dest?, content?, encoding?, mode?, pattern?, recursive? }`
  - data: depende de la operación (contenido, lista de rutas, metadatos, hash SHA-256, etc.)
- Requisitos:
  - Soportar rutas absolutas en Windows (`C:\...`) y relativas al repo.
  - Preservar permisos cuando aplique.
  - `--dry-run` y `--confirm` para borrado/movido destructivo.
  - Tamaño grande: stream para archivos >100MB en lectura/escritura de líneas.
- Casos límite: archivos bloqueados, rutas largas (>260), EOL mixto, codificaciones (UTF-8/UTF-16), nombres con espacios.
- Offline: N/A (local siempre).
- Pruebas mínimas: crear/leer/borrar archivo, glob recursivo, json RW, checksum estable.
- Aceptación: 100% de operaciones básicas pasan en Windows sin dependencias.

---

### 2) Process Runner Tool

- Objetivo: Ejecutar procesos locales, capturar stdout/stderr, código de salida, con timeout y límites de tamaño.
- Idiomas: PowerShell y Python.
- Dependencias: stdlib.
- Contrato:
  - op: `run`
  - args: `{ cmd: string, cwd?: string, env?: object, timeout_ms?: int, shell?: bool }`
  - data: `{ exit_code, stdout, stderr, duration_ms }`
- Requisitos: manejo de códigos no-cero; truncado configurable; redacción de secretos en logs.
- Casos límite: procesos colgados, salida binaria, rutas con espacios, retorno grande.
- Offline: sí (local).
- Pruebas: ejecutar `echo`, comando inexistente, timeout.
- Aceptación: determinista y sin colgar.

---

### 3) HTTP/MOCK Tool

- Objetivo: Cliente HTTP simple con soporte de mocks offline.
- Idiomas: PowerShell (`Invoke-RestMethod`/`Invoke-WebRequest`); Python (`urllib.request`).
- Dependencias: stdlib; sin `requests`.
- Contrato:
  - op: `get|post|put|delete|head`
  - args: `{ url, headers?, body?, timeout_ms?, retries?, cache_ttl_s?, fixture_key? }`
  - data: `{ status, headers, body (string|bytes), cached?: bool, fixture?: bool }`
- Requisitos: caché local opcional en `~/.cache/tooling/http/`; validación de URL; size limits; JSON auto-parse opcional.
- Offline: `--offline` usa `fixtures/http/<fixture_key>.json` o mapea por hash de URL.
- Pruebas: GET a ejemplo público (cuando haya red) y fixture offline.
- Aceptación: rutas offline operativas sin red.

---

### 4) Git Tool (wrapper local)

- Objetivo: Operaciones comunes vía binario `git` si existe; fallback a lectura de `.git` cuando sea posible.
- Idiomas: PowerShell y Python.
- Dependencias: binario `git` opcional; stdlib para fallback.
- Contrato:
  - op: `status|diff|log|add|commit|branch|checkout|rev-parse|ls-files`
  - args: específicos de cada comando.
  - data: salida estructurada (listas, parches, SHAs).
- Requisitos: no requiere red; formatos estables; `--dry-run` para commit (simulado).
- Casos límite: repos sin HEAD, rutas renombradas, CRLF/LF.
- Offline: sí (local); no hace `fetch/pull` en modo offline.
- Pruebas: repo de prueba en `fixtures/git`.
- Aceptación: status/diff/commit simulados operan.

---

### 5) Search Tool (búsqueda/code index)

- Objetivo: grep recursivo por contenido y nombre con regex y glob.
- Idiomas: PowerShell (Select-String) y Python (`re`, `os.walk`).
- Dependencias: stdlib. Soporte opcional a `rg` (ripgrep) si está instalado.
- Contrato:
  - op: `grep|find|list`
  - args: `{ root, pattern, is_regex?: bool, include?: [globs], exclude?: [globs], ignore_git?: bool, multiline?: bool }`
  - data: `{ matches: [{ file, line, col, text }] }`
- Requisitos: respetar `.gitignore` si se solicita; rendimiento en árboles grandes; límite de resultados.
- Casos límite: binarios, archivos enormes, Unicode.
- Offline: sí.
- Pruebas: árbol sintético en `fixtures/search`.
- Aceptación: latencia aceptable (<2s en 10k archivos; objetivo orientativo).

---

### 6) AST/Parser Tool (mínimo viable)

- Objetivo: extraer símbolos (funciones, clases) de Python y JS/TS sin dependencias.
- Idiomas: Python (`ast`), PowerShell con regex simple para JS/TS (MVP).
- Dependencias: stdlib.
- Contrato:
  - op: `list-symbols|refs`
  - args: `{ path, lang?: python|js|ts }`
  - data: `{ symbols: [{ name, kind, line }] }`
- Requisitos: robusto en Python; en JS/TS limitar a heurísticas (sin TypeScript compiler).
- Casos límite: decoradores, clases anidadas, export formas varias.
- Offline: sí.
- Pruebas: archivos de ejemplo.
- Aceptación: cobertura razonable en Python; JS/TS heurístico.

---

### 7) KV Store Tool (SQLite/JSON)

- Objetivo: almacenamiento clave-valor, TTL, namespaces.
- Idiomas: Python (`sqlite3`), PowerShell (JSONL por archivo como fallback).
- Dependencias: stdlib.
- Contrato:
  - op: `get|set|del|keys|purge-expired`
  - args: `{ ns, key?, value?, ttl_s? }`
  - data: `{ value?, keys? }`
- Requisitos: atomicidad en set/del; expiración perezosa; tamaños límite.
- Casos límite: escritura concurrente (usar locking); valores grandes.
- Offline: sí.
- Pruebas: set/get/ttl/expira.
- Aceptación: consistencia básica y sin corrupción.

---

### 8) Cache Tool (FS cache)

- Objetivo: caché de resultados por clave con TTL.
- Idiomas: PowerShell y Python.
- Dependencias: stdlib.
- Contrato: `put|get|del|stats` con `{ key, value?, ttl_s? }`.
- Requisitos: hashing estable (SHA-256), política LRU simple opcional.
- Offline: sí.
- Pruebas: expiración y golpes de caché.
- Aceptación: tasa de aciertos correcta.

---

### 9) Config Tool (.env/JSON/YAML)

- Objetivo: cargar configuración desde env, `.env`, JSON y YAML (sin libs externas: YAML opcional con parser simple o sólo JSON).
- Idiomas: PowerShell y Python.
- Dependencias: stdlib; parser YAML mínimo opcional (propio).
- Contrato: `load` con `{ paths:[], env_prefix?, overrides? }` -> config combinada.
- Requisitos: orden de precedencia claro; no loguear secretos.
- Offline: sí.
- Pruebas: merges y prioridad.
- Aceptación: determinista.

---

### 10) Log Tool (JSON logs)

- Objetivo: logging estructurado JSON con niveles y rotación simple.
- Idiomas: PowerShell y Python.
- Dependencias: stdlib.
- Contrato: `log` con `{ level, msg, fields? }` -> escribe a `logs/agent-YYYYMMDD.log`.
- Requisitos: thread-safe básico; timestamps ISO8601; zona horaria local.
- Offline: sí.
- Pruebas: escritura concurrente leve.
- Aceptación: formato estable.

---

### 11) Scheduler/Queue Tool (file-based)

- Objetivo: colas de trabajos simples y planificador periódico.
- Idiomas: Python (file lock) y PowerShell (semaforo simple).
- Dependencias: stdlib.
- Contrato: `enqueue|dequeue|peek|run-due|list` con `{ queue, job?, run_at? }`.
- Requisitos: consistencia mínima; reintentos; backoff.
- Offline: sí.
- Pruebas: productor/consumidor básico.
- Aceptación: no se pierden jobs en caídas controladas.

---

### 12) Template Tool (plantillas)

- Objetivo: render de plantillas mínimo (sin Jinja). Usar `string.Template` (Python) y `-f` o here-strings (PowerShell) con llaves `${}`.
- Dependencias: stdlib.
- Contrato: `render` con `{ template, vars }`.
- Requisitos: errores cuando faltan claves; no ejecutar código.
- Offline: sí.
- Pruebas: variables, valores por defecto opcionales.
- Aceptación: render correcto y seguro.

---

### 13) Evaluation Harness Tool

- Objetivo: medir latencia, exactitud (string match / regex), diffs y recolectar métricas por lote.
- Idiomas: Python.
- Dependencias: stdlib.
- Contrato: `run` con `{ dataset: [ {id,input,expected} ], fn: external?, metrics:[...] }`.
- Requisitos: reporte CSV/JSON; reproducible; semilla.
- Offline: sí.
- Pruebas: dataset pequeño embebido.
- Aceptación: métricas coherentes.

---

### 14) Tool Adapter (protocolo mínimo)

- Objetivo: estandarizar cómo el agente descubre y llama herramientas locales.
- Diseño: descriptor JSON (`tools/descriptor.json`) con:
  - `name,id,version,bin_path,languages,schema` (ops y contratos)
  - Healthcheck (`--op ping`), `--help`, y `--version`
- Invocación: JSONL por stdin/stdout, 1 operación por línea.
- Offline: flag global `--offline` que propaga a subherramientas.
- Aceptación: listado y ejecución de una operación demo.

---

### 15) Archive Tool (zip/tar)

- Objetivo: comprimir/descomprimir con stdlib (`System.IO.Compression`, `zipfile`).
- Contrato: `zip|unzip|list`.
- Casos límite: rutas largas, permisos.

---

### 16) Diff/Patch Tool

- Objetivo: generar y aplicar parches unificados.
- Idiomas: Python (`difflib`); PowerShell para aplicar con cuidado.
- Contrato: `diff|apply` con `{ base, patch }`.
- Requisitos: back up antes de aplicar; dry-run.

---

### 17) Formatter Tool (opcional)

- Objetivo: formateo mínimo (espacios finales, EOF newline). Si hay `black/prettier` locales, ofrecer wrapper; si no, fallback mínimo.
- Offline: sí.

---

### 18) Secrets Tool (cifrado local)

- Objetivo: guardar/leer secretos cifrados localmente.
- Idiomas: PowerShell usando DPAPI (`System.Security.Cryptography.ProtectedData`), Python `cryptography` NO (externa) -> usar DPAPI vía `ctypes` o sólo PowerShell en MVP.
- Contrato: `set|get|del|list` con `{ key, value? }`.
- Requisitos: sólo local; no subir a VCS; archivo en `%APPDATA%/tooling/secrets.dat`.
- Offline: sí.

## Requisitos no funcionales generales

- Rendimiento: cada operación target < 200 ms local en datasets normales; operaciones intensivas con streaming.
- Trazabilidad: todos emiten logs JSON y aceptan `--trace-id`.
- Ergonomía: mensajes de error claros con `code` y `details`.
- Portabilidad: rutas y EOL compatibles Windows.

## Seguridad

- Sanitizar entradas (paths normales, evitar traversal `..`).
- No registrar secretos.
- Modos `--dry-run` y confirmaciones para operaciones destructivas.

## Pruebas y validación

- Unitarias mínimas (PowerShell: scripts simples; Python: `unittest` stdlib). No usar dependencias externas (Pester opcional más adelante).
- Fixtures en `fixtures/` para reproducibilidad offline.
- Validación manual vía wrapper de comandos del repo, p.ej. colas en `scripts/commands.txt` con ejemplos.

## Roadmap y prioridades (MVP -> Plus)

1. MVP (alta prioridad)
   - FS Tool, Process Runner, Search Tool, Log Tool, Config Tool, HTTP/MOCK con fixtures, Tool Adapter base
2. Plus (media)
   - Git Tool, Cache, KV Store (SQLite/JSON), Template Tool, Archive/Diff
3. Avanzado (baja)
   - Scheduler/Queue, AST/Parser, Formatter, Secrets

## Ejemplos de comandos para el wrapper (documentar en `scripts/commands.txt`)

- FS Tool (crear y leer):
  - `tools\\ps\\fs.ps1 --op write --path d:\\tooling\\tmp.txt --content "hola"`
  - `tools\\ps\\fs.ps1 --op read --path d:\\tooling\\tmp.txt`

- HTTP offline:
  - `tools\\py\\http.py --op get --offline --fixture-key example-users`

- Search:
  - `tools\\ps\\search.ps1 --op grep --root . --pattern "TODO"`

## Referencias y buenas prácticas

- AI Toolkit (AI/Agent): revisar y aplicar guías indicadas por `aitk-get_agent_code_gen_best_practices`, `aitk-get_tracing_code_gen_best_practices`, `aitk-get_evaluation_code_gen_best_practices`, `aitk-evaluation_planner`, `aitk-evaluation_agent_runner_best_practices`.
- Mantener contratos estables y versionados.

## Checklist de aceptación por herramienta (plantilla)

- [ ] CLI y API JSONL implementadas y documentadas
- [ ] `--offline`, `--dry-run`, `--timeout-ms` soportados (si aplica)
- [ ] Logs estructurados a `logs/`
- [ ] Pruebas mínimas pasando (offline)
- [ ] Ejemplos listos para `scripts/commands.txt`

## Siguientes pasos inmediatos

- Crear esqueletos para: FS Tool, Process Runner, Search Tool, Log Tool, Config Tool, HTTP/MOCK Tool, Tool Adapter (PowerShell primero; Python en paralelo).
- Añadir fixtures iniciales en `fixtures/http/` y `fixtures/search/`.
- Documentar comandos MVP en `scripts/commands.txt` con comentarios previos explicando el propósito.
