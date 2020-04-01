const Severity = {
    Normal: 1,
    Warning: 2,
    Critical: 3,
}

function showToast(severity, text) {
    var toast = $('#toast')
    if (!toast.length) {
        $('body').append('<div id="toast" class="toast hidden"></div>')
        toast = $('#toast')
    }

    toast.text(text)

    toast.removeClass('normal')
    toast.removeClass('warning')
    toast.removeClass('critical')

    switch(severity) {
        case Severity.Warning:
            toast.addClass('warning')
            break
        case Severity.Critical:
            toast.addClass('critical')
            break
        case Severity.Normal:
        default:
            toast.addClass('normal')
            break
    }


    toast.removeClass('hidden')
    toast.addClass('visible')
    setTimeout(function() {
        toast.removeClass('visible')
        toast.addClass('hidden')
    }, 2000)
}