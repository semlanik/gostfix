/*
 * MIT License
 *
 * Copyright (c) 2020 Alexey Edelev <semlanik@gmail.com>
 *
 * This file is part of gostfix project https://git.semlanik.org/semlanik/gostfix
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of this
 * software and associated documentation files (the "Software"), to deal in the Software
 * without restriction, including without limitation the rights to use, copy, modify,
 * merge, publish, distribute, sublicense, and/or sell copies of the Software, and
 * to permit persons to whom the Software is furnished to do so, subject to the following
 * conditions:
 *
 * The above copyright notice and this permission notice shall be included in all copies
 * or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED,
 * INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR
 * PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE
 * FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
 * OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 */

var detailsUrl = "details/"

$(document).ready(function(){
    $.ajaxSetup({
        global: false,
        type: "POST"
    })
    $(window).bind('hashchange', requestDetails);
    requestDetails();
    loadStatusLine();
})

function openEmail(id) {
    window.location.hash = detailsUrl + id
}

function requestDetails() {
    var hashLocation = window.location.hash
    if (hashLocation.startsWith("#" + detailsUrl)) {
        var messageId = hashLocation.replace(/^#details\//, "")
        if (messageId != "") {
            $.ajax({
                url: "/messageDetails",
                data: {detailsUrl: messageId},
                success: function(result) {
                    $("#details").html(result);
                    $("#details").show()
                },
                error: function(jqXHR, textStatus, errorThrown) {
                    window.location.hash = ""
                }
            })
        }
    } else {
        $("#details").hide()
    }
}

function closeDetails() {
    window.location.hash = ""
}

function loadStatusLine() {
    $.ajax({
        url: "/statusLine",
        success: function(result) {
            $("#statusLine").html(result);
        },
        error: function(jqXHR, textStatus, errorThrown) {
            //TODO: some toast message here once implemented
        }
    })
}