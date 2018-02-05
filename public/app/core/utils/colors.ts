import _ from 'lodash';
import tinycolor from 'tinycolor2';

export const PALETTE_ROWS = 4;
export const PALETTE_COLUMNS = 14;
export const DEFAULT_ANNOTATION_COLOR = 'rgba(0, 211, 255, 1)';
export const OK_COLOR = 'rgba(11, 237, 50, 1)';
export const ALERTING_COLOR = 'rgba(237, 46, 24, 1)';
export const NO_DATA_COLOR = 'rgba(150, 150, 150, 1)';
export const REGION_FILL_ALPHA = 0.09;

export let colorPalettes = {
  classic: {
    name: 'Classic',
    colors: [
      '#7EB26D',
      '#EAB839',
      '#6ED0E0',
      '#EF843C',
      '#E24D42',
      '#1F78C1',
      '#BA43A9',
      '#705DA0',
      '#508642',
      '#CCA300',
      '#447EBC',
      '#C15C17',
      '#890F02',
      '#0A437C',
      '#6D1F62',
      '#584477',
      '#B7DBAB',
      '#F4D598',
      '#70DBED',
      '#F9BA8F',
      '#F29191',
      '#82B5D8',
      '#E5A8E2',
      '#AEA2E0',
      '#629E51',
      '#E5AC0E',
      '#64B0C8',
      '#E0752D',
      '#BF1B00',
      '#0A50A1',
      '#962D82',
      '#614D93',
      '#9AC48A',
      '#F2C96D',
      '#65C5DB',
      '#F9934E',
      '#EA6460',
      '#5195CE',
      '#D683CE',
      '#806EB7',
      '#3F6833',
      '#967302',
      '#2F575E',
      '#99440A',
      '#58140C',
      '#052B51',
      '#511749',
      '#3F2B5B',
      '#E0F9D7',
      '#FCEACA',
      '#CFFAFF',
      '#F9E2D2',
      '#FCE2DE',
      '#BADFF4',
      '#F9D9F9',
      '#DEDAF7',
    ],
  },
  v5: {
    name: 'v5.0',
    colors: [
      '#5195ce',
      '#629E51',
      '#d683ce', //'#BA43A9',
      '#ea6460', //'#EA6460',
      '#f2c96d', //'#EAB839',
      '#ef843c', //'#6ED0E0',
      '#EF843C',
      '#E24D42',
      '#1F78C1',
      '#705DA0',
      '#508642',
      '#CCA300',
      '#447EBC',
      '#C15C17',
      '#890F02',
      '#0A437C',
      '#6D1F62',
      '#584477',
      '#B7DBAB',
      '#F4D598',
      '#70DBED',
      '#F9BA8F',
      '#F29191',
      '#82B5D8',
      '#E5A8E2',
      '#AEA2E0',
      '#E5AC0E',
      '#64B0C8',
      '#E0752D',
      '#BF1B00',
      '#0A50A1',
      '#962D82',
      '#614D93',
      '#9AC48A',
      '#F2C96D',
      '#65C5DB',
      '#F9934E',
      '#7EB26D',
      '#D683CE',
      '#806EB7',
      '#3F6833',
      '#967302',
      '#2F575E',
      '#99440A',
      '#58140C',
      '#052B51',
      '#511749',
      '#3F2B5B',
      '#E0F9D7',
      '#FCEACA',
      '#CFFAFF',
      '#F9E2D2',
      '#FCE2DE',
      '#BADFF4',
      '#F9D9F9',
      '#DEDAF7',
    ],
  },
};

export function sortColorsByHue(hexColors) {
  let hslColors = _.map(hexColors, hexToHsl);

  let sortedHSLColors = _.sortBy(hslColors, ['h']);
  sortedHSLColors = _.chunk(sortedHSLColors, PALETTE_ROWS);
  sortedHSLColors = _.map(sortedHSLColors, chunk => {
    return _.sortBy(chunk, 'l');
  });
  sortedHSLColors = _.flattenDeep(_.zip(...sortedHSLColors));

  return _.map(sortedHSLColors, hslToHex);
}

export function hexToHsl(color) {
  return tinycolor(color).toHsl();
}

export function hslToHex(color) {
  return tinycolor(color).toHexString();
}

export function getColorPalette(palatte: string) {
  return colorPalettes[palatte].colors;
}

export function getSortedPalette(palatte: string) {
  return sortColorsByHue(getColorPalette(palatte));
}
