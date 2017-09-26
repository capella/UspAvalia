    <script src="<?= $url_full; ?>/assets/js/bootstrap.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/bootstrap-3-typeahead/4.0.2/bootstrap3-typeahead.min.js"></script>
    <script>
    $('input.typeahead').typeahead({
        fitToElement: true,
        minLength: 4,
        source:  function (query, process) {
            return $.post('/search', { query: query }, function (data) {
                data = $.parseJSON(data);
                return process(data);
            });
        },
        matcher: function(item) {
            return true;
        },
        autoSelect: false,
        afterSelect: function (item) {
            if (item.type != 1)
                window.location.href = "/?p=professor&id="+item.id;
            else
                window.location.href = "/?p=disciplina&id="+item.id;
        }
    });
    </script>
    </div><!-- /.container -->
    <title>USP Avalia</title>
  </body>
</html>