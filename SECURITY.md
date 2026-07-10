# Security Policy

## Modelo De Seguridad

EmberDTL modela un servicio custodial de liquidacion con seguros internos por
activo. La capa externa se considera responsable de autenticar comandos; el
motor aplica controles de dominio sobre accounting, limites economicos,
transiciones de estado y reportes deterministas.

## Invariantes Esperadas

- Cada pool de seguro pertenece a un unico activo.
- Una reserve solo acepta depositos del activo configurado.
- Una facility no puede abrirse por encima del saldo de su reserve.
- Un repayment no puede superar el outstanding de la facility.
- Los fees de settlement se separan entre operador y pool de seguro.
- Un default solo se reporta fuera de su ventana de gracia.
- Un claim debe estar asociado a un default activo y a cuentas existentes.
- Un payout de cobertura registra journal, metricas y balance del beneficiario.
- Una recovery resuelta limpia la exposicion pendiente del default.

## Validaciones Automatizadas

La suite de CI ejecuta:

- formato Go mediante `gofmt`;
- `go test ./...`;
- `go vet ./...`;
- build del binario `emberdtl`;
- tests TypeScript de integracion sobre escenarios JSON;
- comprobacion de lineas de codigo de `src/`.

## Gestion De Dependencias

El motor Go no depende de paquetes externos. La capa Node se usa para scripts,
tests de integracion y formato de archivos JSON, Markdown y YAML. Dependabot
revisa modulos Go, dependencias npm y GitHub Actions.

## Alcance De Revision

La revision recomendada cubre:

- admision de facilities;
- accounting de reserves;
- calculo y reparto de fees;
- contribuciones al pool;
- reporte y resolucion de defaults;
- registro y ejecucion de claims;
- reconciliacion por activo.

## Reportes Internos

Los reportes deben incluir:

- escenario JSON minimo;
- salida de `emberdtl run`;
- version de Go y Node;
- impacto economico observado;
- test de regresion propuesto.
