<xsl:stylesheet version="2.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output omit-xml-declaration="yes" indent="yes"/>
    <xsl:template match="node()|@*">
        <xsl:copy>
            <xsl:apply-templates select="node()|@*"/>
        </xsl:copy>
    </xsl:template>
    <xsl:template match="os">
        <os>
            <xsl:apply-templates select="@*|node()"/>
        </os>
    </xsl:template>
    <xsl:template match="/domain/os/loader">
        <xsl:copy>
            <!--<xsl:apply-templates select="node()|@*"/>-->
            <xsl:attribute name="secure">
                <xsl:value-of select="'no'"/>
            </xsl:attribute>
            <xsl:attribute name="readonly">
                <xsl:value-of select="'yes'"/>
            </xsl:attribute>
            <xsl:attribute name="type">
                <xsl:value-of select="'pflash'"/>
            </xsl:attribute>
            <xsl:value-of select="."/>
        </xsl:copy>
    </xsl:template>
    <xsl:template match="/domain/features">
        <xsl:copy>
            <xsl:apply-templates select="node()|@*"/>
            <xsl:element name ="smm" />
        </xsl:copy>
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
