import React from 'react';
import { storiesOf } from '@storybook/react';
import { NamedColorsPalette } from './NamedColorsPalette';
import { getColorName, getColorDefinitionByName } from '../../utils/namedColorsPalette';
import { withKnobs, select } from '@storybook/addon-knobs';
import { withCenteredStory } from '../../utils/storybook/withCenteredStory';
import { UseState } from '../../utils/storybook/UseState';

const BasicGreen = getColorDefinitionByName('green');
const BasicBlue = getColorDefinitionByName('blue');
const LightBlue = getColorDefinitionByName('light-blue');

const NamedColorsPaletteStories = storiesOf('UI/ColorPicker/Palettes/NamedColorsPalette', module);

NamedColorsPaletteStories.addDecorator(withKnobs).addDecorator(withCenteredStory);

NamedColorsPaletteStories.add('Named colors swatch - support for named colors', () => {
  const selectedColor = select(
    'Selected color',
    {
      Green: 'green',
      Red: 'red',
      'Light blue': 'light-blue',
    },
    'red'
  );

  return (
    <UseState initialState={selectedColor}>
      {(selectedColor, updateSelectedColor) => {
        return <NamedColorsPalette color={selectedColor} onChange={updateSelectedColor} />;
      }}
    </UseState>
  );
}).add('Named colors swatch - support for hex values', () => {
  const selectedColor = select(
    'Selected color',
    {
      Green: BasicGreen.variants.dark,
      Red: BasicBlue.variants.dark,
      'Light blue': LightBlue.variants.dark,
    },
    'red'
  );
  return (
    <UseState initialState={selectedColor}>
      {(selectedColor, updateSelectedColor) => {
        return <NamedColorsPalette color={getColorName(selectedColor)} onChange={updateSelectedColor} />;
      }}
    </UseState>
  );
});
