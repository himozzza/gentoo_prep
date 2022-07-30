# Gentoo Prepair chroot!

Этот скрипт скачивает и распаковывает указанную Вами редакцию Gentoo в рабочую директорию /mnt/gentoo.

Доступно четыре редакции amd64:
1. "stage3-amd64-desktop-openrc"
2. "stage3-amd64-desktop-systemd"
3. "stage3-amd64-nomultilib-openrc"
4. "stage3-amd64-nomultilib-systemd"


# ВНИМАНИЕ!!!
Директории /proc /sys /dev /run монтируюся автоматически в новую рабочую область.

Требуется вручную запустить "chroot /mnt/gentoo" и "source /etc/profile. По окончанию это упоминание появится в консоли.



Запуск исключительно из sudo su.
