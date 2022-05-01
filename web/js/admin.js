function updateServiceStats() {
    $.ajax({
        url: '/s',
        success: function(result) {
            var data = jQuery.parseJSON(result);
            pageMax = Math.floor(data.total/50);

            if ($('#mailList')) {
                $('#mailList').html(data.html);
            }
        },
        error: function(jqXHR, textStatus, errorThrown) {
            if ($('#mailList')) {
                $('#mailList').html('Unable to load service list');
            }
        }
    });
}