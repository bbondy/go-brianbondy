(() => {
  const player = document.querySelector('.blog-audio-player');
  const body = document.getElementById('narrated-post');
  const audio = player && player.querySelector('audio');
  if (!player || !body || !audio) return;

  // Match the whitespace normalization used when the narration source is built.
  function normalizedWithPositions(raw) {
    let text = '';
    const positions = [];
    for (let i = 0; i < raw.length; i++) {
      if (/\s/.test(raw[i])) {
        if (text && text[text.length - 1] !== ' ') {
          text += ' ';
          positions.push(i);
        }
      } else {
        text += raw[i];
        positions.push(i);
      }
    }
    return { text: text.trim(), positions };
  }

  function wrapSentence(paragraph, sentence, index, cursor) {
    const raw = paragraph.textContent;
    const normalized = normalizedWithPositions(raw);
    const match = normalized.text.indexOf(sentence.text.replace(/\s+/g, ' ').trim(), cursor);
    if (match < 0) return -1;
    const start = normalized.positions[match];
    const end = normalized.positions[match + sentence.text.length - 1] + 1;
    const walker = document.createTreeWalker(paragraph, NodeFilter.SHOW_TEXT);
    const nodes = [];
    let offset = 0;
    while (walker.nextNode()) {
      const node = walker.currentNode;
      nodes.push({ node, start: offset, end: offset + node.length });
      offset += node.length;
    }
    for (const part of nodes) {
      const from = Math.max(start, part.start) - part.start;
      const to = Math.min(end, part.end) - part.start;
      if (from >= to) continue;
      const fragment = document.createDocumentFragment();
      const content = part.node.textContent;
      fragment.append(document.createTextNode(content.slice(0, from)));
      const span = document.createElement('span');
      span.className = 'audio-sentence';
      span.dataset.audioSentence = String(index);
      span.tabIndex = 0;
      span.setAttribute('role', 'button');
      span.setAttribute('aria-label', `Play from: ${sentence.text}`);
      span.textContent = content.slice(from, to);
      fragment.append(span, document.createTextNode(content.slice(to)));
      part.node.replaceWith(fragment);
    }
    return match + sentence.text.length;
  }

  fetch(player.dataset.timings)
    .then(response => {
      if (!response.ok) throw new Error('Could not load narration timings');
      return response.json();
    })
    .then(data => {
      // Follow the same prose-block order as scripts/blog_audio_source.go.
      const blocks = [];
      const selected = new Set();
      for (const node of body.querySelectorAll('p, h1, h2, h3, h4, h5, h6, li')) {
        if (node.tagName === 'LI' && node.querySelector('p, h1, h2, h3, h4, h5, h6')) continue;
        if (!node.textContent.trim()) continue;
        if ([...selected].some(ancestor => ancestor.contains(node))) continue;
        blocks.push(node);
        selected.add(node);
      }
      const sentences = data.sentences;
      if (!sentences.length) throw new Error('Narration timings are empty');
      const controls = player.querySelector('.blog-audio-controls');
      const toggle = controls.querySelector('.blog-audio-toggle');
      const seek = controls.querySelector('.blog-audio-seek');
      const time = controls.querySelector('.blog-audio-time');
      const expectedDuration = sentences[sentences.length - 1].end + 0.18;

      function duration() {
        return Number.isFinite(audio.duration) && audio.duration > 0 ? audio.duration : expectedDuration;
      }

      function formatTime(seconds) {
        const value = Math.floor(Math.max(0, seconds));
        return `${Math.floor(value / 60)}:${String(value % 60).padStart(2, '0')}`;
      }

      function updateControls() {
        const playing = !audio.paused && !audio.ended;
        toggle.textContent = playing ? '❚❚' : '▶';
        toggle.classList.toggle('is-playing', playing);
        toggle.setAttribute('aria-label', playing ? 'Pause narration' : 'Play narration');
        toggle.title = playing ? 'Pause narration' : 'Play narration';
        seek.value = String(Math.round(1000 * audio.currentTime / duration()));
        time.textContent = `${formatTime(audio.currentTime)} / ${formatTime(duration())}`;
      }

      toggle.addEventListener('click', () => {
        if (audio.paused || audio.ended) audio.play();
        else audio.pause();
      });
      seek.addEventListener('input', () => {
        audio.currentTime = Number(seek.value) * duration() / 1000;
        updateControls();
      });
      for (const event of ['timeupdate', 'seeked', 'play', 'pause', 'ended', 'loadedmetadata', 'durationchange']) {
        audio.addEventListener(event, updateControls);
      }
      updateControls();
      controls.hidden = false;
      player.classList.add('is-enhanced');

      const groups = new Map();
      sentences.forEach((sentence, index) => {
        const blockIndex = sentence.block ?? sentence.paragraph;
        if (!groups.has(blockIndex)) groups.set(blockIndex, []);
        groups.get(blockIndex).push({ sentence, index });
      });

      for (const [blockIndex, group] of groups) {
        const block = blocks[blockIndex];
        if (!block) continue;
        // Work backwards so wrapping text does not change earlier offsets.
        let cursor = 0;
        const normalized = normalizedWithPositions(block.textContent).text;
        const matches = [];
        for (const item of group) {
          const text = item.sentence.text.replace(/\s+/g, ' ').trim();
          const found = normalized.indexOf(text, cursor);
          if (found < 0) break;
          matches.push({ ...item, found });
          cursor = found + text.length;
        }
        if (matches.length !== group.length) continue;
        for (let i = matches.length - 1; i >= 0; i--) {
          wrapSentence(block, matches[i].sentence, matches[i].index, matches[i].found);
        }
      }

      let current = -1;
      function updateHighlight() {
        const time = audio.currentTime;
        let next = -1;
        for (let i = 0; i < sentences.length; i++) {
          if (time >= sentences[i].start && time < sentences[i].end) {
            next = i;
            break;
          }
        }
        if (next === current) return;
        body.querySelectorAll('.audio-sentence.is-current').forEach(node => node.classList.remove('is-current'));
        current = next;
        if (next >= 0) {
          body.querySelectorAll(`[data-audio-sentence="${next}"]`).forEach(node => node.classList.add('is-current'));
        }
      }

      audio.addEventListener('timeupdate', updateHighlight);
      audio.addEventListener('seeked', updateHighlight);
      audio.addEventListener('ended', updateHighlight);
      function seekToSentence(event) {
        if (event.target.closest('a')) return;
        const span = event.target.closest('[data-audio-sentence]');
        if (!span) return;
        audio.currentTime = sentences[Number(span.dataset.audioSentence)].start;
        audio.play();
      }
      body.addEventListener('click', seekToSentence);
      body.addEventListener('keydown', event => {
        if (event.key !== 'Enter' && event.key !== ' ') return;
        if (!event.target.matches('[data-audio-sentence]')) return;
        event.preventDefault();
        seekToSentence(event);
      });
    })
    .catch(error => console.warn('Narration highlighting unavailable:', error));
})();
