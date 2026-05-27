#!/usr/bin/env bash
# rename-module.sh — `templatev27` module-ийн нэрийг шинэ Go module нэр болгож
# нийт repo-д өөрчилнө. Template-ийг clone хийсний дараа эхэлж энэ script-ийг
# нэг л удаа ажиллуулна.
#
# Хэрэглэх жишээ:
#   ./scripts/rename-module.sh github.com/myorg/my-api
#
# Юу хийдэг вэ:
#   1. backend/go.mod-ийн "module templatev27" мөрийг шинэ нэрээр сольно.
#   2. бүх .go файл доторх "templatev27/..." import-ыг шинэ нэр + "/..." болгон солино.
#   3. swagger docs (`backend/docs/`) дотрох static "templatev27" мөрүүдийг солино.
#   4. backend.env.example болон docker-compose.yml дахь template-тусгай service
#      нэрс (gerege-template гэх мэт) утга өөрчлөгдөхгүй — те доорх "грep" гэдэг
#      алхмаар үлдсэн ишлэлүүдийг шалгана.
#
# git тутамд commit хийхгүй — script ажилласны дараа `git diff`-ыг өөрөө шалгаж,
# `git commit`-ыг өөрөө хийнэ үү.

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <new-module-path>" >&2
  echo "  e.g.   $0 github.com/myorg/my-api" >&2
  exit 1
fi

NEW="$1"
OLD="templatev27"

# Эх module-аас өөр модуль нэрийг ялгахын тулд жаахан паттернтэй байя.
# .git, vendor, node_modules доторх файлуудыг тоохгүй.

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

echo "==> Replacing module path: $OLD → $NEW"
echo "==> Repo root: $ROOT"

if ! grep -q "^module $OLD$" backend/go.mod; then
  echo "FATAL: backend/go.mod-д 'module $OLD' мөр олдсонгүй. Аль хэдийн rename хийгдсэн юм болов уу?" >&2
  exit 2
fi

# 1) go.mod
sed -i.bak "s|^module $OLD$|module $NEW|" backend/go.mod
rm -f backend/go.mod.bak

# 2) Go файлуудын import path
#    -path '*/.git/*' -prune-аар .git-ийг тойрно. node_modules / vendor байхгүй.
find backend -type f -name '*.go' -not -path '*/.git/*' -print0 \
  | xargs -0 sed -i.bak "s|\"$OLD/|\"$NEW/|g; s|\"$OLD\"|\"$NEW\"|g"
find backend -type f -name '*.go.bak' -delete

# 3) Swagger docs дахь string mention (godoc annotation-аас үүсдэг)
if [[ -d backend/docs ]]; then
  find backend/docs -type f \( -name '*.go' -o -name '*.json' -o -name '*.yaml' \) -print0 \
    | xargs -0 -r sed -i.bak "s|$OLD|$NEW|g"
  find backend/docs -type f -name '*.bak' -delete
fi

# 4) go.sum-ыг шинэ module path-аар дахин шийдвэрлэх — go.sum нь шууд module
#    нэрийг агуулдаггүй (зөвхөн dependency-ийн нэрс) тул хариу мэдрэх ёсгүй,
#    гэхдээ хэрэглэгчид эцсийн алхмыг сануулъя.

echo
echo "✓ Module path replaced."
echo
echo "Дараа нь:"
echo "  1) cd backend && go mod tidy   # шинэ нэрээр зүгшрүүлэх"
echo "  2) make test                   # тест бүгд гарч ирэхийг шалгах"
echo "  3) git diff                    # өөрчлөлтүүдийг харах"
echo "  4) git add -A && git commit -m 'chore: rename module to $NEW'"
