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
$(document).ready(function(){
    $.ajaxSetup({
        global: false,
        type: "POST"
    })
    $(window).bind('hashchange', requestDetails)
    requestDetails()
    loadStatusLine()
    clearInterval(updateTimerId)
    // updateMessageList()
    updateTimerId = setInterval(updateMessageList, updateInterval)
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
                data: {messageId: messageId},
                success: function(result) {
                    $("#mail"+messageId).removeClass("unread")
                    $("#mail"+messageId).addClass("read")
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

function closeDetails() {
    window.location.hash = ""
}

function loadStatusLine() {
    $.ajax({
        url: "/statusLine",
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

function setRead(messageId, read) {
    $.ajax({
        url: "/setRead",
        data: {messageId: messageId,
               read: read},
        success: function(result) {
            if (read) {
                if ($("#readIcon"+messageId)) {
                    $("#readIcon"+messageId).attr("src", "/assets/read.svg")
                }
                $("#mail"+messageId).removeClass("unread")
                $("#mail"+messageId).addClass("read")
            } else {
                if ($("#readIcon"+messageId)) {
                    $("#readIcon"+messageId).attr("src", "/assets/unread.svg")
                }
                $("#mail"+messageId).removeClass("read")
                $("#mail"+messageId).addClass("unread")
            }
        },
        error: function(jqXHR, textStatus, errorThrown) {
        }
    })
}

function toggleRead(messageId) {
    if ($("#readIcon"+messageId)) {
        setRead(messageId, $("#readIcon"+messageId).attr("src") == "/assets/unread.svg")
    }
}

function removeMail(messageId) {
    $.ajax({
        url: "/remove",
        data: {messageId: messageId},
        success: function(result) {
            $("#mail"+messageId).remove();
            closeDetails()
        },
        error: function(jqXHR, textStatus, errorThrown) {
        }
    })
}

function setDetailsVisible(visible) {
    if (visible) {
        $("#details").show()
        $("#messageList").css({pointerEvents: "none"})
        clearInterval(updateTimerId)
    } else {
        $("#details").hide()
        $("#details").html("")
        $("#messageList").css({pointerEvents: "auto"})
        updateTimerId = setInterval(updateMessageList, updateInterval)
        updateMessageList()
    }
}

function updateMessageList() {
    $.ajax({
        url: "/messageList",
        success: function(result) {
            if($("#messageList")) {
                // console.log("result: " + result)
                $("#messageList").html(result)
            }
        },
        error: function(jqXHR, textStatus, errorThrown) {
            if($("#messageList")) {
                $("#messageList").html(textStatus)
            }
        }
    })
}