// USP Avalia Frontend JavaScript

class USPAvalia {
    constructor() {
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupCSRFToken();
    }

    setupEventListeners() {
        // Rating buttons
        document.querySelectorAll('.rating-btn').forEach(btn => {
            btn.addEventListener('click', (e) => this.handleRating(e));
        });

        // Comment voting
        document.querySelectorAll('.comment-vote-btn').forEach(btn => {
            btn.addEventListener('click', (e) => this.handleCommentVote(e));
        });

        // Search form enhancements
        const searchForm = document.querySelector('#search-form');
        if (searchForm) {
            searchForm.addEventListener('submit', (e) => this.handleSearch(e));
        }

        // Auto-hide alerts
        setTimeout(() => {
            document.querySelectorAll('.alert-dismissible').forEach(alert => {
                const bsAlert = new bootstrap.Alert(alert);
                bsAlert.close();
            });
        }, 5000);
    }

    setupCSRFToken() {
        // Get CSRF token from meta tag or form
        const csrfToken = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content') ||
                         document.querySelector('input[name="gorilla.csrf.Token"]')?.value;
        
        if (csrfToken) {
            this.csrfToken = csrfToken;
        }
    }

    async handleRating(event) {
        event.preventDefault();
        const button = event.target;
        const classProfessorId = button.dataset.classProfessorId;
        const score = button.dataset.score;
        const type = button.dataset.type || '1';

        // Visual feedback
        button.classList.add('loading');
        
        try {
            const requestData = {
                class_professor_id: parseInt(classProfessorId),
                votes: [{
                    type: parseInt(type),
                    score: parseInt(score)
                }]
            };

            const response = await fetch('/vote-batch', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': this.csrfToken
                },
                body: JSON.stringify(requestData)
            });

            if (response.ok) {
                this.showThankYouModal();
                this.updateRatingDisplay(button, score);
            } else {
                throw new Error('Erro ao registrar voto');
            }
        } catch (error) {
            this.showMessage('Erro ao registrar voto. Tente novamente.', 'danger');
        } finally {
            button.classList.remove('loading');
        }
    }

    async handleCommentVote(event) {
        event.preventDefault();
        const button = event.target;
        const commentId = button.dataset.commentId;
        const vote = button.dataset.vote;

        button.classList.add('loading');

        try {
            const response = await fetch('/vote-comment', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded',
                },
                body: new URLSearchParams({
                    'comment_id': commentId,
                    'vote': vote,
                    'gorilla.csrf.Token': this.csrfToken
                })
            });

            if (response.ok) {
                this.showMessage('Voto no comentário registrado!', 'success');
                this.updateCommentVoteDisplay(button, vote);
            } else {
                throw new Error('Erro ao votar no comentário');
            }
        } catch (error) {
            this.showMessage('Erro ao votar no comentário. Tente novamente.', 'danger');
        } finally {
            button.classList.remove('loading');
        }
    }

    updateRatingDisplay(button, score) {
        // Update button appearance
        const container = button.closest('.rating-buttons');
        if (container) {
            container.querySelectorAll('.btn').forEach(btn => {
                btn.classList.remove('btn-primary', 'btn-outline-primary');
                btn.classList.add('btn-outline-primary');
            });
            
            button.classList.remove('btn-outline-primary');
            button.classList.add('btn-primary');
        }

        // Disable other buttons temporarily
        container.querySelectorAll('.btn').forEach(btn => {
            btn.disabled = true;
        });

        setTimeout(() => {
            container.querySelectorAll('.btn').forEach(btn => {
                btn.disabled = false;
            });
        }, 2000);
    }

    updateCommentVoteDisplay(button, vote) {
        const container = button.closest('.comment-votes');
        if (container) {
            const upBtn = container.querySelector('[data-vote="1"]');
            const downBtn = container.querySelector('[data-vote="-1"]');
            
            // Reset buttons
            upBtn.classList.remove('btn-success', 'btn-outline-success');
            downBtn.classList.remove('btn-danger', 'btn-outline-danger');
            
            if (vote === '1') {
                upBtn.classList.add('btn-success');
                downBtn.classList.add('btn-outline-danger');
            } else {
                upBtn.classList.add('btn-outline-success');
                downBtn.classList.add('btn-danger');
            }
        }
    }

    handleSearch(event) {
        const input = event.target.querySelector('input[name="q"]');
        const query = input.value.trim();
        
        if (!query) {
            event.preventDefault();
            this.showMessage('Digite algo para pesquisar', 'warning');
            input.focus();
        }
    }

    showThankYouModal() {
        // Check if modal exists
        const modal = document.getElementById('thankYouModal');
        if (modal) {
            // Use Bootstrap 3 modal syntax
            $('#thankYouModal').modal('show');
        }
    }

    showMessage(message, type = 'info') {
        // Remove existing alerts
        document.querySelectorAll('.alert-floating').forEach(alert => alert.remove());
        
        const alertDiv = document.createElement('div');
        alertDiv.className = `alert alert-${type} alert-dismissible fade show alert-floating`;
        alertDiv.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        
        document.body.appendChild(alertDiv);
        
        // Auto-hide after 3 seconds
        setTimeout(() => {
            if (alertDiv.parentNode) {
                const bsAlert = new bootstrap.Alert(alertDiv);
                bsAlert.close();
            }
        }, 3000);
    }

    // Utility functions
    debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    new USPAvalia();
});

// Export for testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = USPAvalia;
}
