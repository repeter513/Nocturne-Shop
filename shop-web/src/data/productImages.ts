export const productImages: Record<string, string> = {
  '1': 'https://images.unsplash.com/photo-1527864550417-7fd91fc51a46?w=800&q=80',
  '2': 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&q=80',
  '3': 'https://images.unsplash.com/photo-1544947950-fa07a98d237f?w=800&q=80',
  '4': 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=800&q=80',
  '5': 'https://images.unsplash.com/photo-1617806118233-18e1de247200?w=800&q=80',
  '6': 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=800&q=80',
  '7': 'https://images.unsplash.com/photo-1625948515291-69613efd103f?w=800&q=80',
  '8': 'https://images.unsplash.com/photo-1498050108023-c5249f4df085?w=800&q=80',
  '9': 'https://images.unsplash.com/photo-1532012197267-da84d127e765?w=800&q=80',
  '10': 'https://images.unsplash.com/photo-1481627834876-b7833e8f5570?w=800&q=80',
  '11': 'https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=800&q=80',
  '12': 'https://images.unsplash.com/photo-1496181133206-80ce9b88a853?w=800&q=80',
  '13': 'https://images.unsplash.com/photo-1517336714731-489689fd1ca8?w=800&q=80',
  '14': 'https://images.unsplash.com/photo-1618384887929-16ec33fab9ef?w=800&q=80',
  '15': 'https://images.unsplash.com/photo-1572569511254-d8f925fe2cbb?w=800&q=80',
}

export function getProductImage(id: string): string | undefined {
  return productImages[id]
}

export function getShortDescription(description: string): string {
  return description.split('\n\n')[0] ?? description
}

export function getDetailParagraphs(description: string): string[] {
  return description.split('\n\n').filter(Boolean)
}
