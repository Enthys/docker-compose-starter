const logsContainer = document.getElementById("logs");
for (let i = 0; i < 1000; i++) {
	const log = document.createElement('p');
	log.innerText = i;
	log.classList.add('log')
	logsContainer.prepend(log)
}

