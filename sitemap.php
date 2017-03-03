<?php require_once('Connections/CapellaResumo.php');
header("Content-type: text/xml");
mysql_select_db($database_CapellaResumo, $CapellaResumo);
$query_Paginas = "SELECT id FROM aulaprofessor";
$Paginas = mysql_query($query_Paginas, $CapellaResumo) or die(mysql_error());
$row_Paginas = mysql_fetch_assoc($Paginas);
$totalRows_Paginas = mysql_num_rows($Paginas);

header("Content-type: text/xml");
?>
<?= '<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" 
  xmlns:image="http://www.google.com/schemas/sitemap-image/1.1" 
  xmlns:video="http://www.google.com/schemas/sitemap-video/1.1">'; ?>
              <url>
                <loc> http://uspavalia.com/ </loc>
            </url> 
              <url>
                <loc> http://uspavalia.com/?p=email </loc>
            </url>
              <url>
                <loc> http://uspavalia.com/?p=sobre </loc>
            </url> 
<?php do { ?>
            <url>
                <loc> http://uspavalia.com/?p=ver&amp;id<?= $row_Paginas['id'];?> </loc>
            </url>      
<?php } while ($row_Paginas = mysql_fetch_assoc($Paginas)); ?>
</urlset>
<?php
mysql_free_result($Paginas);
?>