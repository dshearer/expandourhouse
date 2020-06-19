const gFirstCongressStartYear = 1789;

function congressStartYear(congress) {
    return gFirstCongressStartYear + 2*(congress-1);
}

export {congressStartYear};