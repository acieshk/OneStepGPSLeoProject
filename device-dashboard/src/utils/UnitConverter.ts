export const convertUnits = (value: number, 
	fromUnit: string, 
	toUnit: string): number => {
	const conversionFactors: { [key: string]: { [key: string]: number } } = {
		'km': {
			'mi': 0.621371,
			'm': 1000,
			'km': 1
		},
		'mi': {
			'km': 1.60934,
			'm': 1609.34,
			'mi': 1
		},
		'm': {
			'km': 0.001,
			'mi': 0.000621371,
			'm': 1
		},
		'km/h': {
			'mph': 0.621371,
			'km/h': 1
		},
		'mph': {
			'km/h': 1.60934,
			'mph': 1
		}

	};

	if (!conversionFactors[fromUnit] || !conversionFactors[fromUnit][toUnit]) {
		console.warn(`Conversion from ${fromUnit} to ${toUnit} not supported.`);
		return value; // Return original value if conversion is not supported
	}

	return value * conversionFactors[fromUnit][toUnit];
};
