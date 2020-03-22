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
var mailbox = ""
var pageMax = 10
const mailboxRegex = /^(\/m\d+)/g
const emailRegex = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
const emailEndRegex = /[;,\s]/g

var folders = new Array()
var notifierSocket = null

var toEmailList = new Array()
var toEmailIndex = 0
var toEmailPreviousSelectionPosition = 0

$(window).click(function(e){
    var target = $(e.target)
    var isDropDown = false
    for (var i = 0; i < target.parents().length; i++) {
        isDropDown = target.parents()[i].classList.contains("dropbtn")
        if (isDropDown) {
            break
        }
    }
    if (!e.target.matches('.dropbtn') && !isDropDown) {
        $(".dropdown-content").hide()
    }
})

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

    $("#mailNewButton").click(mailNew)
    connectNotifier()

    $("#toEmailField").on("input", toEmailFieldChanged)
    $("#toEmailField").keyup(function(e){
        if (e.keyCode == 8 && toEmailPreviousSelectionPosition == 0 && e.target.selectionStart == 0 && toEmailList.length > 0 && $("#toEmailList").children().length > 0) {
            removeToEmail($("#toEmailList").children().last().attr("id"), toEmailList[toEmailList.length - 1])
        }
        toEmailPreviousSelectionPosition = e.target.selectionStart
    })
})

function toEmailFieldChanged(e) {
    const selectionPosition = e.target.selectionStart - 1
    var actualText = $("#toEmailField").val()

    if (actualText.length <= 0 || selectionPosition <= 0) {
        return
    }

    var lastChar = actualText[selectionPosition]
    if (lastChar.match(emailEndRegex)) {
        var toEmail = actualText.slice(0, selectionPosition)
        $("#toEmailField").val(actualText.slice(selectionPosition + 1, actualText.length))
        if (toEmail.length <= 0) {
            return
        }
        var style = toEmail.match(emailRegex) ? "valid" : "invalid"
        $("#toEmailList").append("<div class=\""+ style + " toEmail\" id=\"toEmail" + toEmailIndex + "\">" + toEmail + "<img class=\"iconBtn\" style=\"height: 12px; margin-left:10px; margin: auto;\" onclick=\"removeToEmail('toEmail" + toEmailIndex + "', '" + toEmail + "');\" src=\"/assets/cross.svg\"/></div>")
        toEmailIndex++
        toEmailList.push(toEmail)
    }
}

function removeToEmail(id, email) {
    const index = toEmailList.indexOf(email)
    if (index >= 0) {
        toEmailList.splice(index, 1)
    }

    $("#" + id).remove()
    console.log("Remove email: " + email + " index:" + index)
    console.log("toEmailList: " + toEmailList)
}

function mailNew(e) {
    window.location.hash = currentFolder + currentPage + "/mailNew"
}

function mailOpen(id) {
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
    page = 0
    if (hashParts.length >= 3 && hashParts[2] != "") {
        page = parseInt(hashParts[2])
        if (typeof page != "number" || page > pageMax || page < 0) {
            page = 0
        }
    }

    if (hashParts.length >= 2 && (hashParts[1] != currentFolder || currentPage != page) && hashParts[1] != "") {
        updateMailList(hashParts[1], page)
    }

    if (hashParts.length >= 4 && hashParts[3] != "" && hashParts[3] != "/mailNew") {
        if (currentMail != hashParts[3]) {
            requestMail(hashParts[3])
        }
    } else {
        setDetailsVisible(false)
    }

    hashParts = hashLocation.split("/")
    if (hashParts.length == 2 && hashParts[1] == "mailNew") {
        setMailNewVisible(true)
    } else {
        setMailNewVisible(false)
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
                folderStat(currentFolder);//TODO: receive statistic from websocket
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
            folderList = jQuery.parseJSON(result)
            for(var i = 0; i < folderList.folders.length; i++) {
                folders.push(folderList.folders[i].name)
                folderStat(folderList.folders[i].name)
            }
            $("#folders").html(folderList.html)
        },
        error: function(jqXHR, textStatus, errorThrown) {
            //TODO: some toast message here once implemented
        }
    })
}

function folderStat(folder) {
    $.ajax({
        url: mailbox + "/folderStat",
        data: {
            folder: folder
        },
        success: function(result) {
            var stats = jQuery.parseJSON(result)
            if (stats.unread > 0) {
                $("#folderStats"+folder).text(stats.unread)
                $("#folder"+folder).addClass("unread")
            } else {
                $("#folder"+folder).removeClass("unread")
                $("#folderStats"+folder).text("")
            }
        },
        error: function(jqXHR, textStatus, errorThrown) {
            //TODO: some toast message here once implemented
        }
    })
}

function closeDetails() {
    window.location.hash = currentFolder + currentPage
}

function closeMailNew() {
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

function localDate(elementToChange, timestamp) {
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

    $("#"+elementToChange).text(dateString)
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
                if ($("#readListIcon"+mailId)) {
                    $("#readListIcon"+mailId).attr("src", "/assets/read.svg")
                }
                $("#mail"+mailId).removeClass("unread")
                $("#mail"+mailId).addClass("read")
            } else {
                if ($("#readIcon"+mailId)) {
                    $("#readIcon"+mailId).attr("src", "/assets/unread.svg")
                }
                if ($("#readListIcon"+mailId)) {
                    $("#readListIcon"+mailId).attr("src", "/assets/unread.svg")
                }
                $("#mail"+mailId).removeClass("read")
                $("#mail"+mailId).addClass("unread")
            }
            folderStat(currentFolder);//TODO: receive statistic from websocket
        },
        error: function(jqXHR, textStatus, errorThrown) {
        }
    })
}

function toggleRead(mailId, iconId) {
    if ($("#"+iconId+mailId)) {
        setRead(mailId, $("#"+iconId+mailId).attr("src") == "/assets/unread.svg")
    }
}

function removeMail(mailId, callback) {
    var url = currentFolder != "Trash" ? "/remove" : "/delete"
    $.ajax({
        url: url,
        data: {mailId: mailId},
        success: function(result) {
            $("#mail"+mailId).remove();
            if (callback) {
                callback();
            }
            folderStat(currentFolder);//TODO: receive statistic from websocket
            folderStat("Trash");//TODO: receive statistic from websocket
        },
        error: function(jqXHR, textStatus, errorThrown) {
        }
    })
}


function restoreMail(mailId, callback) {
    var url = "/restore"
    $.ajax({
        url: url,
        data: {mailId: mailId},
        success: function(result) {
            if (currentFolder == "Trash") {
                $("#mail"+mailId).remove();
            }
            if (callback) {
                callback();
            }
            for (var i = 0; i < folders.length; i++) {
                folderStat(folders[i])
            }
        },
        error: function(jqXHR, textStatus, errorThrown) {
        }
    })
}

function setDetailsVisible(visible) {
    if (visible) {
        $("#mailDetails").show()
        $("#mailList").css({pointerEvents: "none"})
    } else {
        currentMail = ""
        $("#mailDetails").hide()
        $("#mailDetails").html("")
        $("#mailList").css({pointerEvents: "auto"})
    }
}

function setMailNewVisible(visible) {
    if (visible) {
        $("#mailNew").show()
        $("#mailList").css({pointerEvents: "none"})
    } else {
        currentMail = ""
        $("#mailNew").hide()
        $("#mailList").css({pointerEvents: "auto"})
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
            pageMax = Math.floor(data.total/50)

            if ($("#mailList")) {
                $("#mailList").html(data.html)
            }
            currentFolder = folder
            currentPage = page

            if($("#currentPageIndex")) {
                $("#currentPageIndex").text(currentPage + 1)
            }
            if($("#totalPageCount")) {
                $("#totalPageCount").text(pageMax + 1)
            }
        },
        error: function(jqXHR, textStatus, errorThrown) {
            if ($("#mailList")) {
                $("#mailList").html("Unable to load message list")
            }
        }
    })
}

function nextPage() {
    var newPage = currentPage < (pageMax - 1) ? currentPage + 1 : pageMax
    window.location.hash = currentFolder + newPage
}

function prevPage() {
    var newPage = currentPage > 0 ? currentPage - 1 : 0
    window.location.hash = currentFolder + newPage
}

function toggleDropDown(dd) {
    $("#"+dd).toggle()
}

function sendNewMail() {
    var composedEmailString = toEmailList[0]
    for(var i = 1; i < toEmailList.length; i++) {
        composedEmailString += "," + toEmailList[i]
    }
    $("#newMailTo").val(composedEmailString)
    var formValue = $("#mailNewForm").serialize()
    console.log("formValue: " + formValue)
    $.ajax({
        url: mailbox + "/sendNewMail",
        data: formValue,
        success: function(result) {
            $("#newMailEditor").val("")
            $("#newMailSubject").val("")
            $("#newMailTo").val("")
            closeMailNew()
        },
        error: function(jqXHR, textStatus, errorThrown) {
            //TODO: some toast message here once implemented
        }
    })
}

function logout() {
    window.location.href = "/logout"
}

function connectNotifier() {
    if (notifierSocket != null) {
        return
    }

    var protocol = "wss://"
    if(window.location.protocol  !== "https:") {
        protocol = "ws://"
    }
    notifierSocket = new WebSocket(protocol + window.location.host + mailbox + "/notifierSubscribe")
    notifierSocket.onopen = function() {
    };
    notifierSocket.onmessage = function (e) {
        for (var i = 0; i < folders.length; i++) {
            folderStat(folders[i])
        }
        updateMailList(currentFolder, currentPage)
    }
    notifierSocket.onclose = function () {
    }
}

window.onbeforeunload = function() {
    if (notifierSocket != null) {
        notifierSocket.onclose = function () {}; // disable onclose handler first
        notifierSocket.close();
    }
};