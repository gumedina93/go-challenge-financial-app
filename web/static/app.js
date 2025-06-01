class ChatApp {
    constructor() {
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;

        this.messagesContainer = document.getElementById('messages');
        this.messageInput = document.getElementById('messageInput');
        this.sendButton = document.getElementById('sendButton');
        this.connectionStatus = document.getElementById('connectionStatus');
        this.statusIndicator = this.connectionStatus.querySelector('.status-indicator');
        this.statusText = this.connectionStatus.querySelector('.status-text');

        this.currentUser = document.querySelector('.chat-header .user-info strong').textContent;

        this.initializeEventListeners();
        this.connect();
    }

    initializeEventListeners() {
        this.sendButton.addEventListener('click', () => this.sendMessage());

        this.messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        // Auto-resize input and limit length
        this.messageInput.addEventListener('input', (e) => {
            if (e.target.value.length > 500) {
                e.target.value = e.target.value.substring(0, 500);
            }
        });
    }

    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;

        this.updateConnectionStatus('connecting', 'Connecting...');

        try {
            this.ws = new WebSocket(wsUrl);

            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.reconnectAttempts = 0;
                this.updateConnectionStatus('connected', 'Connected');
                this.sendButton.disabled = false;
                this.messageInput.disabled = false;
            };

            this.ws.onmessage = (event) => {
                const message = JSON.parse(event.data);
                this.displayMessage(message);
            };

            this.ws.onclose = (event) => {
                console.log('WebSocket closed:', event.code, event.reason);
                this.updateConnectionStatus('disconnected', 'Disconnected');
                this.sendButton.disabled = true;
                this.messageInput.disabled = true;

                if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
                    this.scheduleReconnect();
                }
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.updateConnectionStatus('disconnected', 'Connection Error');
            };

        } catch (error) {
            console.error('Failed to create WebSocket:', error);
            this.updateConnectionStatus('disconnected', 'Connection Failed');
            this.scheduleReconnect();
        }
    }

    scheduleReconnect() {
        this.reconnectAttempts++;
        const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

        this.updateConnectionStatus('connecting', `Reconnecting in ${Math.ceil(delay/1000)}s...`);

        setTimeout(() => {
            if (this.reconnectAttempts <= this.maxReconnectAttempts) {
                this.connect();
            } else {
                this.updateConnectionStatus('disconnected', 'Connection Failed');
            }
        }, delay);
    }

    updateConnectionStatus(status, text) {
        this.statusIndicator.className = `status-indicator ${status}`;
        this.statusText.textContent = text;
    }

    sendMessage() {
        if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
            console.log('WebSocket not connected');
            return;
        }

        const content = this.messageInput.value.trim();
        if (!content) return;

        const message = {
            type: 'message',
            content: content,
            time: new Date().toISOString()
        };

        this.ws.send(JSON.stringify(message));
        this.messageInput.value = '';
        this.messageInput.focus();
    }

    displayMessage(message) {
        const messageElement = document.createElement('div');
        messageElement.className = 'message';

        // Determine message type
        if (message.username === this.currentUser) {
            messageElement.classList.add('own');
        } else if (message.username === 'StockBot') {
            messageElement.classList.add('bot');
        } else {
            messageElement.classList.add('other');
        }

        const time = new Date(message.time);
        const timeString = time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

        messageElement.innerHTML = `
            <div class="message-header">${message.username}</div>
            <div class="message-content">${this.escapeHtml(message.content)}</div>
            <div class="message-time">${timeString}</div>
        `;

        this.messagesContainer.appendChild(messageElement);
        this.scrollToBottom();

        // Keep only last 50 messages in DOM
        while (this.messagesContainer.children.length > 50) {
            this.messagesContainer.removeChild(this.messagesContainer.firstChild);
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    scrollToBottom() {
        this.messagesContainer.scrollTop = this.messagesContainer.scrollHeight;
    }
}

// Initialize the chat app when the page loads
document.addEventListener('DOMContentLoaded', () => {
    new ChatApp();
});