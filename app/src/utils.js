// turns 1 into '1st', etc.
function ordinal(number) {
    if (number < 0) {
      throw 'Ordinal got negative number';
    }
    const suffixes = ['th', 'st', 'nd', 'rd', 'th', 'th', 'th', 'th', 'th', 'th'];
    var suffix;
    if (((number % 100) === 11) || ((number % 100) === 12) || ((number % 100) === 13)) {
      suffix = suffixes[0];
    }
    else {
      suffix = suffixes[number % 10];
    }
    return `${number}${suffix}`;
}

export {ordinal};