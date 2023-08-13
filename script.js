const editor = document.getElementById("editor");
const ws = new WebSocket("ws://localhost:8080/streamer");
let previousContent = editor.innerText;
let clientId;

function computeDiff(oldStr, newStr){
    let position = 0;
    while (position < oldStr.length && oldStr[position] === newStr[position]){
        position++;
    }
    if(position === oldStr.length && position === newStr.length){
        return null;
    }

    let oldSuffixStart = oldStr.length - 1;
    let newSuffixStart = newStr.length - 1;

    while(oldSuffixStart >= position && newSuffixStart >= position && oldStr[oldSuffixStart] === newStr[newSuffixStart]){
        oldSuffixStart--;
        newSuffixStart--;
    }
    return {
        position: position,
        oldText: oldStr.slice(position, oldSuffixStart + 1),
        newText: newStr.slice(position, newSuffixStart + 1)
    };
}
function applyChangeToContent(content, change) {
    return content.slice(0, change.position) + 
           change.newText + 
           content.slice(change.position + change.oldText.length);
}
editor.addEventListener("input", (event) => {
    const currentContent = editor.innerText;
    const diff = computeDiff(previousContent, currentContent);
    if(diff){
        const change = {
            type: "text-change",
            position: diff.position,
            oldText: diff.oldText,
            newText: diff.newText,
            clientId: clientId
        };

        ws.send(JSON.stringify(change));
    }

    previousContent = currentContent; // Update for the next input event.
});

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);

    if (data.type === "assign_client_id") {
        clientId = data.clientId;
    } else if(data.type == "text-change"){
        const updatedContent = applyChangeToContent(editor.innerText, data);
        editor.innerText = updatedContent;
        previousContent = updatedContent; // update the previousContent to the latest version
    }else if(data.type === "current_state") {
        editor.innerText = data.content;
        previousContent = data.content; // set the previousContent to the current state
    }
};
