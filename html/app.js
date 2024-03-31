$(document).ready(function () {
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
    console.log('Connected to the WebSocket server');
};

ws.onmessage = (event) => {
    try {
        const messageData = JSON.parse(event.data);
        if (messageData.Date && messageData.Subject && messageData.From && messageData.Message) {
            displayMessage(messageData);
        }
    } catch (e) {
        console.error('Error parsing message data', e);
    }
};

function displayMessage(data) {
    const messagesDiv = jQuery('#messages');
    const messageElement = jQuery('<div></div>').addClass('message');
    
    const messageHTML = `
    <table><tr><td style="width: 90px;">
        <strong>Date:</strong></td><td>${data.Date}</td></tr><tr><td>
        <strong>From:</strong></td><td>${data.From}</td></tr><tr><td>
        <strong>Subject:</strong></td><td>${data.Subject}</td></tr>
    </table>
    <hr>
        <div class="message-content">${data.Message}</div>
    `;
    
    messageElement.html(messageHTML);
    messagesDiv.prepend(messageElement);

    const hideButton = jQuery('<button>Hide</button>');
    hideButton.addClass("button_done");
    hideButton.on('click', function() {
        messageElement.slideUp();
    });
    messageElement.append(hideButton);

    const buttonSeparator = jQuery('<span> </span>');
    messageElement.append(buttonSeparator);


    // Toggle button functionality
    const toggleButton = jQuery('<button>Original message</button>'); // Initially set to "Summary"
    toggleButton.addClass("button_done");
    let showingOriginal = false;
    toggleButton.on('click', function() {
        const messageContent = messageElement.find('.message-content');
        messageContent.fadeOut(function() {
            if (showingOriginal) {
                messageContent.html(data.Message);
                toggleButton.text('Original message'); // Change button text to "Summary"
            } else {
                messageContent.html(data.Original);
                toggleButton.text('Summary'); // Change button text to "Original"
            }
            showingOriginal = !showingOriginal;
            messageContent.fadeIn();
        });
    });
    messageElement.append(toggleButton);
    $(this).scrollTop(0);
}
});