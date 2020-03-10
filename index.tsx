import { red, blue, green } from "https://deno.land/std/fmt/colors.ts";

enum Direction {
  RISE = 'RISE',
  FALL = 'FALL',
}

type Item = {
  date: Date,
  direction: Direction
}

const parseFile = async (filepath: string): Promise<Item[]> => {
  const decoder = new TextDecoder('utf-8')
  const data = await Deno.readFile(filepath)

  const lines = decoder.decode(data).split('\n')

  const stream = lines.map((line) => {
    const match = line.match(/(\d{4}\/\d{2}\/\d{2} \d{2}:\d{2}:\d{2}\.\d{6}) ([-+]{3})/)
    if (match == null) {
      return undefined
    }
    const date = new Date(match[1])
    const direction = match[2] === '+++' ? Direction.RISE : Direction.FALL

    return { date, direction } as Item
  }).filter(Boolean)

  return stream as Item[]
}

const detectStream = (stream: Item[]) => {
  let previousDate = null
  let state = null

  let sum = 0

  for (const item of stream) {
    if (previousDate != null && state !== item.direction) {
      const difference = item.date.getTime() - previousDate.getTime()

      console.log(blue((Math.round(difference / 100)/10).toFixed(1)), green(`${state} --> ${item.direction}`))

      if (difference / 1000 > 0.5 && item.direction === Direction.FALL) {
        console.log(red(item.date.toString()), ++sum)
      }
    }

    previousDate = item.date
    state = item.direction
  }
}

const stream = await parseFile('/home/admin/Downloads/gastly_main-03.03.20_07_58_33_(+1300).txt')

detectStream(stream)
