const contactIcons = {
  discord: require('@/assets/images/groups/contacts/discord.png'),
  vk: require('@/assets/images/groups/contacts/vk.png'),
  telegram: require('@/assets/images/groups/contacts/telegram.png'),
  twitch: require('@/assets/images/groups/contacts/twitch.png'),
  youtube: require('@/assets/images/groups/contacts/youtube.png'),
  whatsapp: require('@/assets/images/groups/contacts/whatsapp.png'),
  max: require('@/assets/images/groups/contacts/max.png'),
  default: require('@/assets/images/groups/contacts/default.png'),
};

export const getContactIconByLink = (link: string) => {
  if (!link) return contactIcons.default;
  
  const lowerLink = link.toLowerCase();

  if (lowerLink.includes('discord.gg') || lowerLink.includes('discord.com')) {
    return contactIcons.discord;
  }
  if (lowerLink.includes('vk.com') || lowerLink.includes('vk.ru')) {
    return contactIcons.vk;
  }
  if (lowerLink.includes('t.me') || lowerLink.includes('telegram')) {
    return contactIcons.telegram;
  }
  if (lowerLink.includes('twitch.tv')) {
    return contactIcons.twitch;
  }
  if (lowerLink.includes('youtube.com') || lowerLink.includes('youtu.be')) {
    return contactIcons.youtube;
  }
  if (lowerLink.includes('wa.me') || lowerLink.includes('whatsapp')) {
    return contactIcons.whatsapp;
  }
  if (lowerLink.includes('max.ru')) {
    return contactIcons.max;
  }
  
  return contactIcons.default;
};