function injectReporter() {
    let oldParse = JSON.parse;
    JSON.parse = function newParse(d, r) {
        let data = oldParse(d, r);
        setTimeout(function() {
        if (data.d && data.d.b && data.d.b.d && data.d.b.d.type == 'rollresult') {
            $.post('http://127.0.0.1:1323/roll', JSON.stringify(data.d.b));
        }}, 0);
        return data;
    };
}

function getPlayers() {
    var players = {};
    $('.by').each((i,r) => players[$(r).parent().attr('data-playerid')] = $(r).text().slice(0, -1));
    console.log(players);
    $.post('http://127.0.0.1:1323/players', JSON.stringify(players));
}

getPlayers();
injectReporter();
console.info('DiceSpy injected.');
