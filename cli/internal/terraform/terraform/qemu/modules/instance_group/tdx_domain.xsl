<xsl:stylesheet version="2.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform" xmlns:qemu='http://libvirt.org/schemas/domain/qemu/1.0'>
  <xsl:output omit-xml-declaration="yes" indent="yes"/>
    <xsl:template match="node()|@*">
        <xsl:copy>
            <xsl:apply-templates select="node()|@*"/>
        </xsl:copy>
    </xsl:template>
    <xsl:template match="/domain">
        <xsl:copy>
            <xsl:apply-templates select="node()|@*"/>
            <xsl:element name ="clock">
                <xsl:attribute name="offset">
                    <xsl:value-of select="'utc'"/>
                </xsl:attribute>
                <xsl:element name ="timer">
                    <xsl:attribute name="name">
                        <xsl:value-of select="'hpet'"/>
                    </xsl:attribute>
                    <xsl:attribute name="present">
                        <xsl:value-of select="'no'"/>
                    </xsl:attribute>
                </xsl:element>
            </xsl:element>
            <xsl:element name ="on_poweroff"><xsl:text>destroy</xsl:text></xsl:element>
            <xsl:element name ="on_reboot"><xsl:text>restart</xsl:text></xsl:element>
            <xsl:element name ="on_crash"><xsl:text>destroy</xsl:text></xsl:element>
            <xsl:element name ="pm">
                <xsl:element name ="suspend-to-mem">
                     <xsl:attribute name="enable">
                        <xsl:value-of select="'no'"/>
                    </xsl:attribute>
                </xsl:element>
                <xsl:element name ="suspend-to-disk">
                    <xsl:attribute name="enable">
                        <xsl:value-of select="'no'"/>
                    </xsl:attribute>
                </xsl:element>
            </xsl:element>
            <xsl:element name ="allowReboot">
                <xsl:attribute name="value">
                    <xsl:value-of select="'no'"/>
                </xsl:attribute>
            </xsl:element>
            <xsl:element name ="launchSecurity">
                <xsl:attribute name="type">
                    <xsl:value-of select="'tdx'"/>
                </xsl:attribute>
                <xsl:element name ="policy"><xsl:text>0x10000001</xsl:text></xsl:element>
                <xsl:element name ="Quote-Generation-Service"><xsl:text>vsock:2:4050</xsl:text></xsl:element>
            </xsl:element>
            <xsl:element name ="qemu:commandline" >
                <xsl:element name ="qemu:arg">
                    <xsl:attribute name="value">
                        <xsl:value-of select="'-cpu'"/>
                    </xsl:attribute>
                </xsl:element>
                <xsl:element name ="qemu:arg">
                     <xsl:attribute name="value">
                        <xsl:value-of select="'host,-kvm-steal-time'"/>
                    </xsl:attribute>
                </xsl:element>
            </xsl:element>
        </xsl:copy>
    </xsl:template>
    <xsl:template match="os">
        <os>
            <xsl:apply-templates select="@*|node()"/>
        </os>
    </xsl:template>
    <xsl:template match="/domain/os/loader">
        <loader>
            <xsl:apply-templates select="node()"/>
        </loader>
    </xsl:template>
    <xsl:template match="/domain/features">
        <features>
            <acpi/>
            <apic/>
            <ioapic driver="qemu"/>
        </features>
    </xsl:template>
    <xsl:template match="/domain/vcpu">
        <vcpu placement="static"><xsl:apply-templates select="@*|node()"/></vcpu>
    </xsl:template>
    <xsl:template match="/domain/cpu">
        <xsl:copy>
            <xsl:apply-templates select="node()|@*"/>
            <xsl:element name ="topology">
                <xsl:attribute name="sockets">
                 <xsl:value-of select="'1'"/>
                </xsl:attribute>
                <xsl:attribute name="cores">
                 <xsl:value-of select="'1'"/>
                </xsl:attribute>
                <xsl:attribute name="threads">
                 <xsl:value-of select="'1'"/>
                </xsl:attribute>
            </xsl:element>
        </xsl:copy>
    </xsl:template>
    <xsl:template match="/domain/devices/console">
        <console type="pty">
            <target type="virtio" port="1" />
      </console>
    </xsl:template>
    <xsl:template match="/domain/devices/graphics"></xsl:template>
    <xsl:template match="/domain/devices/rng"></xsl:template>
</xsl:stylesheet>
