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
var updateTimerId = null
var updateInterval = 5000
var mailbox = ""
var mailboxRegex = /^(\/m\d+)/g

$(document).ready(function(){
    $.ajaxSetup({
        global: false,
        type: "POST"
    })
    $(window).bind('hashchange', requestDetails)
    requestDetails()

    urlPaths = mailboxRegex.exec($(location).attr('pathname'))
    if (urlPaths != null && urlPaths.length === 2) {
        mailbox = urlPaths[0]
    } else {
        mailbox = ""
    }

    loadFolders()
    loadStatusLine()
    updateMailList()
    if(mailbox != "") {
        clearInterval(updateTimerId)
        updateTimerId = setInterval(updateMailList, updateInterval)
    }
})

function openEmail(id) {
    window.location.hash = detailsUrl + id
}

function requestDetails() {
    var hashLocation = window.location.hash
    if (hashLocation.startsWith("#" + detailsUrl)) {
        var mailId = hashLocation.replace(/^#details\//, "")
        if (mailId != "") {
            $.ajax({
                url: "/mail",
                data: {mailId: mailId},
                success: function(result) {
                    $("#mail"+mailId).removeClass("unread")
                    $("#mail"+mailId).addClass("read")
                    $("#details").html(result);
                    setDetailsVisible(true);
                },
                error: function(jqXHR, textStatus, errorThrown) {
                    $("#details").html(textStatus)
                    setDetailsVisible(true)
                }
            })
        }
    } else {
        setDetailsVisible(false)
    }
}

function loadFolders() {
    if (mailbox == "") {
        $("#folders").html("Unable to load folder list")
        return
    }

    $.ajax({
        url: mailbox + "/folders",
        success: function(result) {
            $("#folders").html(result)
        },
        error: function(jqXHR, textStatus, errorThrown) {
            //TODO: some toast message here once implemented
        }
    })
}

function closeDetails() {
    window.location.hash = ""
}

function loadStatusLine() {
    $.ajax({
        url: mailbox + "/statusLine",
        success: function(result) {
            $("#statusLine").html(result)
        },
        error: function(jqXHR, textStatus, errorThrown) {
            //TODO: some toast message here once implemented
        }
    })
}

function localDate(timestamp) {
    var today = new Date()
    var date = new Date(timestamp*1000)

    dateString = ""
    if (today.getDay() == date.getDay()
        && today.getMonth() == date.getMonth()
        && today.getFullYear() == date.getFullYear()) {
            dateString = date.toLocaleTimeString("en-US")
    } else if (today.getFullYear() == date.getFullYear()) {
        const options = { day: 'numeric', month: 'short' }
        dateString = date.toLocaleDateString("en-US", options)
    } else {
        dateString = date.toLocaleDateString("en-US")
    }

    return dateString
}

function setRead(mailId, read) {
    $.ajax({
        url: "/setRead",
        data: {mailId: mailId,
               read: read},
        success: function(result) {
            if (read) {
                if ($("#readIcon"+mailId)) {
                    $("#readIcon"+mailId).attr("src", "/assets/read.svg")
                }
                $("#mail"+mailId).removeClass("unread")
                $("#mail"+mailId).addClass("read")
            } else {
                if ($("#readIcon"+mailId)) {
                    $("#readIcon"+mailId).attr("src", "/assets/unread.svg")
                }
                $("#mail"+mailId).removeClass("read")
                $("#mail"+mailId).addClass("unread")
            }
        },
        error: function(jqXHR, textStatus, errorThrown) {
        }
    })
}

function toggleRead(mailId) {
    if ($("#readIcon"+mailId)) {
        setRead(mailId, $("#readIcon"+mailId).attr("src") == "/assets/unread.svg")
    }
}

function removeMail(mailId) {
    $.ajax({
        url: "/remove",
        data: {mailId: mailId},
        success: function(result) {
            $("#mail"+mailId).remove();
            closeDetails()
        },
        error: function(jqXHR, textStatus, errorThrown) {
        }
    })
}

function setDetailsVisible(visible) {
    if (visible) {
        $("#details").show()
        $("#mailList").css({pointerEvents: "none"})
        clearInterval(updateTimerId)
    } else {
        $("#details").hide()
        $("#details").html("")
        $("#mailList").css({pointerEvents: "auto"})
        updateTimerId = setInterval(updateMailList, updateInterval)
    }
}

function updateMailList() {
    if (mailbox == "") {
        if($("#mailList")) {
            $("#mailList").html("Unable to load message list")
        }
        return
    }

    $.ajax({
        url: mailbox + "/mailList",
        success: function(result) {
            if($("#mailList")) {
                // console.log("result: " + result)
                $("#mailList").html(result)
            }
        },
        error: function(jqXHR, textStatus, errorThrown) {
            if($("#mailList")) {
                $("#mailList").html(textStatus)
            }
        }
    })
}