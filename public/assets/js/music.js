var gameModal = new bootstrap.Modal(document.getElementById('modal-game'))

$(document).ready(function() {
    $('#song-guess').autocomplete({ // Jquery UI autocomplete
        source: AllSongs
    })
})

function animateProgressBar(duration, targetProgress) { // Animates the progress bar to show song timestamp
    const progressBar = document.getElementById('progressBar');
    const progressBarParentWidth = progressBar.parentNode.getBoundingClientRect().width;
    const startWidth = 0;
    const targetWidth = progressBarParentWidth * targetProgress / 100;
    
    progressBar.style.transition = 'none';
    progressBar.style.width = `${startWidth}px`;
    
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        progressBar.style.transition = `width ${duration}ms linear`;
        progressBar.style.width = `${targetWidth}px`;
      });
    });
  }
  
let audio; // Audio object to play the song
let currDuration = 0.5 // Current duration of the song snippet
let durations = [0.5, 1, 3, 5, 15, 30] // Possible durations of the song snippet
var widths = { 0.5: 2, 1: 5, 3: 8, 5: 25, 15: 50, 30: 100 } // Maps each duration to the width of the progress bar

let currProgressStatus = 1; // The current status element to show the result of each turn, 1-6

$('#btn-play-audio').on('click', function(e) { 
    e.preventDefault();
    if(audio) { // If audio has already been received, then play it
        audio.pause()
        audio.currentTime = 0
        audio.play();
        animateProgressBar(currDuration * 1000, widths[currDuration]) 
        return;
    }
    fetch('http://localhost:8080/audio').then(res => res.blob()).then(blob => { // Else fetch the audio from the server
        if(audio) {
            audio.pause();
        }
        const url = window.URL.createObjectURL(new Blob([blob])); 
        audio = new Audio(url);

        audio.play();

        animateProgressBar(currDuration * 1000, widths[currDuration])        
    })
})


let awaitingSkipResponse = false; // Boolean to check if the user is awaiting a response from the server
$('#btn-skip').on('click', function(e) { // Skip to the next song snippet length
    e.preventDefault();
    if(awaitingSkipResponse) return; // If the user is awaiting a response, then do not send another request
    awaitingSkipResponse = true;
    document.getElementById('btn-skip').disabled = true; // Disable the skip button
    fetch('http://localhost:8080/skip').then(res => res.blob()).then(blob => { // Fetch the next song snippet from server
        document.getElementById('btn-skip').disabled = false; // Enable the skip button
        awaitingSkipResponse = false
        if(audio) { // If audio is currently playing, pause it
            audio.pause();
        }
        const url = window.URL.createObjectURL(new Blob([blob]));
        audio = new Audio(url);
        
        audio.play(); // Play the new audio

        currDuration = durations[durations.indexOf(currDuration) + 1] ? durations[durations.indexOf(currDuration) + 1] : currDuration // Update the current duration to new length of song snippet
        animateProgressBar(currDuration * 1000, widths[currDuration])

        if(currProgressStatus >= 6) { // If user has skipped more than 6 times, then end the game
            fetch('http://localhost:8080/finish', { // Send a request to the server to end the game, will shut down session and return the song name
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                    },
                }).then(res => res.json()).then(data => {
                    document.getElementById('game-status').innerText = `You did not guess the song`
                    document.getElementById('song-name').innerText = `${data.song}`
                    gameModal.show() // Show the modal with the song name
                })
            return;
        };
        document.getElementById(`progress-status-${currProgressStatus}`).value = "Skipped" // Update the status element to show that the user skipped
        currProgressStatus += 1
        
    })
})

let awaitingResponse = false; // Boolean to check if the user is awaiting a response from the server

$('#guess-song').on('submit', function(e) { 
    e.preventDefault();
    if(awaitingResponse) return; // If the user is awaiting a response, then do not send another request
    const guess = $('#song-guess').val()
    if(AllSongs.indexOf(guess) < 0) { // If the user's guess is not in the list of songs, then don't submit
        document.getElementById('song-guess').classList.add('is-invalid')
        $('#song-guess').val('')
        setTimeout(() => {
            document.getElementById('song-guess').classList.remove('is-invalid')
        }, 3000)
    }
    else {
        awaitingResponse = true; // Set awaiting response to true
        fetch('http://localhost:8080/guess', { // Send a request to the server to check if the guess is correct
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ guess })
        }).then(res => res.json()).then(data => {
            awaitingResponse = false; // Set awaiting response to false
            if(data.correct) { // If the guess is correct, then end the game
                document.getElementById('song-guess').classList.add('is-valid')

                document.getElementById(`progress-status-${currProgressStatus}`).value = guess
                document.getElementById(`progress-status-${currProgressStatus}`).classList.add('is-valid')

                document.getElementById('game-status').innerHTML = `You <span style="color: green">correctly</span> guessed the song`
                document.getElementById('song-name').innerText = `${guess}`
                gameModal.show()
            }
            else if(data.correct === false) { // If the guess is incorrect, then update the status element to show the guess
                document.getElementById('song-guess').classList.add('is-invalid')

                document.getElementById(`progress-status-${currProgressStatus}`).value = guess
                document.getElementById(`progress-status-${currProgressStatus}`).classList.add('is-invalid')
                currProgressStatus += 1
                $('#song-guess').val('')
                setTimeout(() => {
                    document.getElementById('song-guess').classList.remove('is-invalid')
                }, 3000)
            }
            else { // If user ran out of turns, or final guess was incorrect, then end the game
                fetch('http://localhost:8080/finish', { // Send a request to the server to end the game, will shut down session and return the song name
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                        },
                    }).then(res => res.json()).then(data => {
                        document.getElementById('game-status').innerHTML = `You did <span style="color: red">not</span> guess the song`
                        document.getElementById('song-name').innerText = `${data.song}`
                        gameModal.show() // Show the modal with the song name
                    })
                return;
            }
        })
    }
})



