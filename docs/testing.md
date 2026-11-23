# Проверка перед публикацией

GitHub Actions запускает `go test ./...`, `go vet ./...` и проверку JavaScript. Docker Buildx собирает образы ARMv5, ARMv6, ARMv7, ARM64 и AMD64. Каждый архив проверяется импортом в Docker. Артефакт `routeros-installation-files` содержит эти архивы, установщик и контрольные суммы. Это сборки для тестирования с доступом через GitHub Actions.

Испытания на RouterOS проверяют полный путь установки. Начинайте с CHR и одного клиентского IP, затем повторяйте сценарии на физических моделях. Статусы и критерии приведены в [таблице совместимости](compatibility.md).

Перед публичным релизом загрузите проверенные файлы как отдельные assets одного GitHub Release. Для установки по ссылке используется адрес этого тега. Убедитесь, что внешний пользователь может скачать скрипт и нужный архив по HTTPS, выполните [установку по ссылке](installation.md#установить-с-https-ссылки) на чистом RouterOS и перенесите результаты в таблицу. При подготовке дистрибутива учитывайте условия включённого Mihomo и списка доменов V2Fly. [Mihomo](https://github.com/MetaCubeX/mihomo) · [V2Fly domain-list-community](https://github.com/v2fly/domain-list-community) · [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases)
