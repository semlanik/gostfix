const passwordRegex = /[A-Z0-6!"\#$%&'()*+,\-./:;<=>?@\[\\\]^_‘{|}~]/
const fullNameRegex = /^[\w]+[\w ]*$/

function initControls() {
    $('.inpt, .password').find('.icon').mousedown(function(e) {
        if ($(e.target).parent().hasClass('password')) {
            $(e.target).parent().find('input').attr('type', 'text')
        }
    })

    $('.inpt, .password').find('.icon').mouseup(function(e) {
        if ($(e.target).parent().hasClass('password')) {
            $(e.target).parent().find('input').attr('type', 'password')
        }
    })
}

function addValidation(field, form, func) {
    func(field)
    $(field).on('input', function(e) {func(e.target, form)})
}

function validateForm(form) {
    if (form == null) {
        return false
    }

    if ($(form).find('.inpt').hasClass('bad')) {
        $(form).find('.btn').addClass('disabled')
        return false
    }
    $(form).find('.btn').removeClass('disabled')
    return true
}

function validatePassword(name, form) {
    var element = $(name)
    var fieldDiv = element.parent()
    if (element.val() != '') {
        fieldDiv.removeClass('bad')
        if (element.val().length < 8 || !passwordRegex.test(element.val())) {
            fieldDiv.addClass('weak')
        } else {
            fieldDiv.removeClass('weak')
        }
    } else {
        fieldDiv.removeClass('weak')
        fieldDiv.addClass('bad')
    }
    validateForm(form)
}

function validateFullName(name, form) {
    var element = $(name)
    var fieldDiv = element.parent()
    if (fullNameRegex.test(element.val())) {
        fieldDiv.removeClass('bad')
    } else {
        fieldDiv.addClass('bad')
    }
    validateForm(form)
}

function validateField(element, form) {
    var fieldDiv = element.parent()
    if (element.val() != '') {
        fieldDiv.removeClass('bad')
    } else {
        fieldDiv.addClass('bad')
    }
    validateForm(form)
}

//Email validation
const emailRegex = /^[a-zA-Z]+[\w\d\._-]*$/
var emailInputTimer = null
function validateEmail(name, form) {
    var element = $(name)
    var fieldDiv = element.parent()
    fieldDiv.addClass('bad')
    if (emailRegex.test(element.val())) {
        clearTimeout(emailInputTimer)
        emailInputTimer = setTimeout(function(){
                $.ajax({
                url: '/checkEmail',
                data: {
                    user: element.val()
                },
                success: function(result) {
                    fieldDiv.removeClass('bad')
                    validateForm(form)
                },
                error: function(jqXHR, textStatus, errorThrown) {
                    fieldDiv.addClass('bad')
                    validateForm(form)
                }
            })
        }, 200)
        return
    } else {
        validateForm(form)
    }
}
