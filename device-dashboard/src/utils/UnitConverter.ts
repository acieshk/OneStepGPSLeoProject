type Unit = 'km' | 'mi' | 'm' | 'km/h' | 'mph';

export const convertUnits = (value: number, fromUnit: Unit, toUnit: Unit): number => {
    const conversionFactors: Record<Unit, Record<Unit, number | undefined>> = {
        'km': {
            'mi': 0.621371,
            'm': 1000,
            'km': 1,
        },
        'mi': {
            'km': 1.60934,
            'm': 1609.34,
            'mi': 1,
        },
        'm': {
            'km': 0.001,
            'mi': 0.000621371,
            'm': 1,
        },
        'km/h': {
            'mph': 0.621371,
            'km/h': 1,
        },
        'mph': {
            'km/h': 1.60934,
            'mph': 1,
        },
    };

    const factor = conversionFactors[fromUnit]?.[toUnit];

    if (factor === undefined) {
        throw new Error(`Conversion from ${fromUnit} to ${toUnit} is not supported.`); // Or return NaN
    }

    return value * factor;
};