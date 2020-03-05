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

var currentFolder = ""
var currentPage = 0
var currentMail = ""
var updateTimerId = null
var updateInterval = 50000
var mailbox = ""
var pageMax = 10
const mailboxRegex = /^(\/m\d+)/g

$(document).ready(function(){
    $.ajaxSetup({
        global: false,
        type: "POST"
    })

    urlPaths = mailboxRegex.exec($(location).attr('pathname'))
    if (urlPaths != null && urlPaths.length === 2) {
        mailbox = urlPaths[0]
    } else {
        mailbox = ""
    }

    $(window).bind('hashchange', onHashChanged)
    onHashChanged()
    loadFolders()
    loadStatusLine()

    if (mailbox != "") {
        clearInterval(updateTimerId)
        updateTimerId = setInterval(updateMailList, updateInterval, currentFolder+currentPage)
    }
})

function openEmail(id) {
    window.location.hash = currentFolder + currentPage + "/" + id
}

function openFolder(folder) {
    window.location.hash = folder
}

function onHashChanged() {
    var hashLocation = window.location.hash
    if (hashLocation == "") {
        setDetailsVisible(false)
        openFolder("Inbox")
        return
    }

    hashRegex = /^#([a-zA-Z]+)(\d*)\/?([A-Fa-f\d]*)/g
    hashParts = hashRegex.exec(hashLocation)
    console.log("Hash parts: " + hashParts)
    page = 0
    if (hashParts.length >= 3 && hashParts[2] != "") {
        console.log("page found: " + hashParts[2])
        page = hashParts[2]
    }

    if (hashParts.length >= 2 && (hashParts[1] != currentFolder || currentPage != page) && hashParts[1] != "") {

        updateMailList(hashParts[1], page)
    }

    if (hashParts.length >= 4 && hashParts[3] != "") {
        if (currentMail != hashParts[3]) {
            requestMail(hashParts[3])
        }
    } else {
        setDetailsVisible(false)
    }


}

function requestMail(mailId) {
    if (mailId != "") {
        $.ajax({
            url: "/mail",
            data: {
                mailId: mailId
            },
            success: function(result) {
                currentMail = mailId
                $("#mail"+mailId).removeClass("unread")
                $("#mail"+mailId).addClass("read")
                $("#mailDetails").html(result);
                setDetailsVisible(true);
            },
            error: function(jqXHR, textStatus, errorThrown) {
                $("#mailDetails").html(textStatus)
                setDetailsVisible(true)
            }
        })
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
    window.location.hash = currentFolder + currentPage
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
        $("#mailDetails").show()
        $("#mailList").css({pointerEvents: "none"})
        clearInterval(updateTimerId)
    } else {
        currentMail = ""
        $("#mailDetails").hide()
        $("#mailDetails").html("")
        $("#mailList").css({pointerEvents: "auto"})
        updateTimerId = setInterval(updateMailList, updateInterval, currentFolder+currentPage)
    }
}

function updateMailList(folder, page) {
    if (mailbox == "" || folder == "") {
        if ($("#mailList")) {
            $("#mailList").html("Unable to load message list")
        }
        return
    }

    $.ajax({
        url: mailbox + "/mailList",
        data: {
            folder: folder,
            page: page
        },
        success: function(result) {
            var data = jQuery.parseJSON(result)
            var mailCount = data.total
            if ($("#mailList")) {
                $("#mailList").html(data.html)
            }
            currentFolder = folder
            currentPage = page
        },
        error: function(jqXHR, textStatus, errorThrown) {
            if ($("#mailList")) {
                $("#mailList").html("Unable to load message list")
            }
        }
    })
}

function nextPage() {
    var newPage = currentPage > 0 ? currentPage - 1 : 0
    window.location.hash = currentFolder + newPage
}

function prevPage() {
    var newPage = currentPage < (pageMax - 1) ? currentPage + 1 : pageMax
    window.location.hash = currentFolder + newPage
}