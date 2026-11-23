# Установка MikroTik VPN

Сначала включите контейнеры в RouterOS. Затем выберите архив по [таблице совместимости](compatibility.md). Сборки для проверки доступны в артефактах GitHub Actions после успешного запуска проверки. Скачанный из Actions ZIP нужно распаковать на компьютере.

## Включить контейнеры в RouterOS

1. Откройте Terminal в WinBox и выполните команды ниже. Запишите `version` и `architecture-name` из первого вывода. Если во втором выводе уже есть активный пакет `container`, переходите к шагу 3.

   ```routeros
   /system/resource/print
   /system/package/print where name=container
   ```

2. На сайте MikroTik скачайте **Extra Packages** для той же версии RouterOS и архитектуры. Для hAP ac³ выбирайте `arm`. Распакуйте ZIP, положите подходящий файл `container-*.npk` в корень **Files** в WinBox и перезагрузите роутер командой `/system/reboot`. После перезагрузки повторите `/system/package/print where name=container`. Для CHR пакет выбирается на странице загрузок CHR. [Загрузки MikroTik](https://mikrotik.com/download) · [Управление пакетами](https://manual.mikrotik.com/docs/cli-reference/system/package/)

3. Проверьте режим устройства. Если в выводе указано `container: no`, запустите изменение и подтвердите его на устройстве в течение пяти минут.

   ```routeros
   /system/device-mode/print
   /system/device-mode/update container=yes
   ```

   На физическом MikroTik можно кратко отключить и снова подать питание. Применяйте способ подтверждения, указанный в терминале для модели. На CHR остановите и снова запустите виртуальную машину. После запуска снова выполните `/system/device-mode/print` и убедитесь, что указано `container: yes`. [Режим устройства RouterOS](https://manual.mikrotik.com/docs/system-information-and-utilities/device-mode/)

4. Подключите накопитель и проверьте его командой `/disk/print detail`. Для hAP ac³ используйте USB-накопитель с ext4. Для загрузки и распаковки образа вместе держите примерно 200 МиБ свободного места. Установщик показывает подходящие смонтированные диски и свободный объём. [Руководство MikroTik по контейнерам](https://manual.mikrotik.com/docs/containers/) · [Управление дисками](https://manual.mikrotik.com/docs/hardware/disks/)

## Установить из локального архива

Загрузите `mikrotik-vpn-install.rsc` в корень **Files**, а архив образа — на USB-диск. В интерактивном терминале выполните:

```routeros
/import file-name=mikrotik-vpn-install.rsc
```

Enter принимает предложенное значение. Для сети контейнера предлагается `172.30.55.0/30`. При необходимости введите свой IPv4 CIDR. Укажите существующие списки интерфейсов LAN и WAN. Для пробного запуска укажите один IPv4-адрес клиента, например `192.168.88.50`. При пустом поле клиента установится панель, а маршрутизация клиентов останется выключенной. Установщик предложит сгенерировать пароль или принять введённый вручную.

Для стандартной подсети откройте `http://172.30.55.2:8080` из доверенной сети. Вход выполняется под именем `admin`. Добавьте подписку, проверьте статус узлов и задайте первое правило через VPN. Проверьте также прямой маршрут с того же клиента.

## Установить с HTTPS-ссылки

На HTTPS-сервере разложите рядом скрипт и архивы под их штатными именами. Для GitHub Releases используйте путь конкретного тега, когда релиз и его файлы станут общедоступны. Роутеру нужны DNS, доступ к серверу и доверенная цепочка TLS. Выполните весь блок в интерактивном терминале, подставив адрес каталога вместо примера:

```routeros
{
    :local releaseUrl "https://downloads.example.org/mikrotik-vpn/v0.1.0"
    :local installerFile ("mikrotik-vpn-" . [:rndstr length=8] . ".rsc")
    :global mikrotikVpnReleaseBase
    :onerror err in={
        /tool/fetch url=($releaseUrl . "/mikrotik-vpn-install.rsc") dst-path=$installerFile check-certificate=yes http-max-redirect-count=5
        :set mikrotikVpnReleaseBase $releaseUrl
        /import file-name=$installerFile
    } do={
        :set mikrotikVpnReleaseBase
        :do { /file/remove [/file/find where name=$installerFile] } on-error={}
        :error $err
    }
    /file/remove [/file/find where name=$installerFile]
}
```

При нажатии Enter в поле архива установщик скачает файл для обнаруженной архитектуры. Можно ввести другой путь или HTTPS-ссылку. Образ загружается на выбранный диск, и архив остаётся там после установки. [RouterOS Fetch](https://manual.mikrotik.com/docs/cli-reference/tool/fetch/) · [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases)
