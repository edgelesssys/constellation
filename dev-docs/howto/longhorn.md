# Longhorn on Constellatioin

To build Longhorn compatible images, apply the following changes. Those stem from [their installation guide](https://longhorn.io/docs/1.3.2/deploy/install/#installation-requirements).

```diff
diff --git a/image/mkosi.conf.d/azure.conf b/image/mkosi.conf.d/azure.conf
index bc4b707b..6de2254a 100644
--- a/image/mkosi.conf.d/azure.conf
+++ b/image/mkosi.conf.d/azure.conf
@@ -1,3 +1,5 @@
 [Content]
 Packages=
            WALinuxAgent-udev
+           nfs-utils
+           iscsi-initiator-utils
diff --git a/image/mkosi.skeleton/etc/fstab b/image/mkosi.skeleton/etc/fstab
index e22f0b24..2e212267 100644
--- a/image/mkosi.skeleton/etc/fstab
+++ b/image/mkosi.skeleton/etc/fstab
@@ -1,5 +1,6 @@
 /dev/mapper/state        /run/state         ext4    defaults,x-systemd.makefs,x-mount.mkdir    0    0
 /run/state/var           /var               none    defaults,bind,x-mount.mkdir                0    0
+/run/state/iscsi         /etc/iscsi         none    defaults,bind,x-mount.mkdir                0    0
 /run/state/kubernetes    /etc/kubernetes    none    defaults,bind,x-mount.mkdir                0    0
 /run/state/etccni        /etc/cni/          none    defaults,bind,x-mount.mkdir                0    0
 /run/state/opt           /opt               none    defaults,bind,x-mount.mkdir                0    0
diff --git a/image/mkosi.skeleton/usr/lib/systemd/system-preset/30-constellation.preset b/image/mkosi.skeleton/usr/lib/systemd/system-preset/30-constellation.preset
index 24072c48..7b7498d6 100644
--- a/image/mkosi.skeleton/usr/lib/systemd/system-preset/30-constellation.preset
+++ b/image/mkosi.skeleton/usr/lib/systemd/system-preset/30-constellation.preset
@@ -4,3 +4,5 @@ enable containerd.service
 enable kubelet.service
 enable systemd-networkd.service
 enable tpm-pcrs.service
+enable iscsibefore.service
+enable iscsid.service
diff --git a/image/mkosi.skeleton/usr/lib/systemd/system/iscsibefore.service b/image/mkosi.skeleton/usr/lib/systemd/system/iscsibefore.service
new file mode 100644
index 00000000..355a2f83
--- /dev/null
+++ b/image/mkosi.skeleton/usr/lib/systemd/system/iscsibefore.service
@@ -0,0 +1,12 @@
+[Unit]
+Description=before iscsid
+Before=iscsid.service
+ConditionPathExists=!/etc/iscsi/initiatorname.iscsi
+
+[Service]
+Type=oneshot
+ExecStart=/bin/bash -c "echo \"InitiatorName=$(/sbin/iscsi-iname)\" > /etc/iscsi/initiatorname.iscsi"
+RemainAfterExit=yes
+
+[Install]
+WantedBy=multi-user.target
```
