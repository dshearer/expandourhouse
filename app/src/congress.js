const gFirstCongressStartYear = 1789;

function startYearForCongress(congress) {
    return gFirstCongressStartYear + 2*(congress-1);
}

function congressForYear(year) {
    if (year < gFirstCongressStartYear) {
        return undefined;
    }
	if (year%2 === 0) {
		--year;
	}
	return (year-1789)/2 + 1;
}

export {startYearForCongress, congressForYear};