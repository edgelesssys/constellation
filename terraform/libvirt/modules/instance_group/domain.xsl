<xsl:stylesheet version="2.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output omit-xml-declaration="yes" indent="yes"/>
    <xsl:template match="node()|@*">
        <xsl:copy>
            <xsl:apply-templates select="node()|@*"/>
        </xsl:copy>
    </xsl:template>
    <xsl:template match="os">
        <os firmware="efi">
            <xsl:apply-templates select="@*|node()"/>
        </os>
    </xsl:template>
    <xsl:template match="/domain/devices/tpm/backend">
    <xsl:copy>
        <xsl:apply-templates select="node()|@*"/>
        <xsl:element name ="active_pcr_banks">
            <xsl:element name="sha1"></xsl:element>
            <xsl:element name="sha256"></xsl:element>
            <xsl:element name="sha384"></xsl:element>
            <xsl:element name="sha512"></xsl:element>
        </xsl:element>
    </xsl:copy>
  </xsl:template>
</xsl:stylesheet>
