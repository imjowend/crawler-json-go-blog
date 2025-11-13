# Scraper del Blog de Go

Este script descarga los primeros 25 posts del blog oficial de Go y los guarda en formato JSON.

## Requisitos

- Go 1.16 o superior
- Conexión a Internet

## Uso

```bash
go run main.go
```

El script:
1. Descargará cada uno de los 25 posts más antiguos del blog de Go
2. Extraerá título, fecha, URL y contenido de cada post
3. Mostrará el progreso en consola: "Procesando post X/25..."
4. Guardará todo en `posts.json`

## Formato del JSON resultante

```json
[
  {
    "id": 1,
    "title": "Go: What's New in March 2010",
    "date": "2010-03-18",
    "url": "https://go.dev/blog/go-whats-new-march-2010",
    "content": "Texto completo del post..."
  },
  ...
]
```

## Características

- ✅ Solo usa librerías estándar de Go
- ✅ Manejo robusto de errores
- ✅ Headers HTTP apropiados para evitar bloqueos
- ✅ Pausa entre peticiones para no sobrecargar el servidor
- ✅ Limpieza automática de HTML a texto plano
- ✅ Normalización de fechas a formato YYYY-MM-DD

## Nota

Si ejecutas el script desde un entorno con restricciones de red, puedes recibir errores 403. Ejecuta el script en tu máquina local para mejores resultados.
